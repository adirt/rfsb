// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package server

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/adirt/rfsb/protos"
	"github.com/adirt/rfsb/server/handlers"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	port = 50051
)

type rfsbServer struct {
	browser *handlers.Browser
	fetcher *handlers.Fetcher
}

// type pbMessage interface {
// 	Reset()
// 	String() string
// 	ProtoMessage()
// 	Descriptor()
// }

// type requestHandler interface {
// 	HandleRequest() *pbMessage
// }

func NewServer() (*rfsbServer, error) {
	browser, err := handlers.CreateBrowser()
	if err != nil {
		log.Fatalf("failed to initialize browse handler: %v", err)
	}
	fetcher, err := handlers.CreateFetcher()
	if err != nil {
		log.Fatalf("failed to initialize fetch handler: %v", err)
	}
	return &rfsbServer{browser: browser, fetcher: fetcher}, nil
}

func (this *rfsbServer) Serve() error {
	grpcServer := grpc.NewServer()
	pb.RegisterRemoteFileSystemBrowserServer(grpcServer, this)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.New("failed to listen: " + err.Error())
	}
	if err := grpcServer.Serve(listener); err != nil {
		return errors.New("failed to serve: " + err.Error())
	}
	return nil
}

func (this *rfsbServer) Browse(ctx context.Context, request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	log.Printf("Received browse request for %s", request.Dir)
	response, err := this.browser.HandleRequest(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (this *rfsbServer) Fetch(request *pb.FetchRequest, stream pb.RemoteFileSystemBrowser_FetchServer) error {
	log.Printf("Received fetch request")
	fileCount := 0
	streamChannel := make(chan *pb.FetchResponse)
	fileCountChannel := make(chan int)
	doneChannel := make(chan bool, 1)

	go this.fetcher.HandleRequest(request, streamChannel, fileCountChannel, doneChannel)
	for {
		select {
		case fileCountUpdate := <-fileCountChannel:
			fileCount += fileCountUpdate
			if fileCount == 0 {
				log.Printf("Done streaming requested data")
				doneChannel <- true
				return nil
			}
		case fileChunk := <-streamChannel:
			printChunkInfo(fileChunk)
			if err := stream.Send(fileChunk); err != nil {
				return err
			}
		}
	}
}

func printChunkInfo(fileChunk *pb.FetchResponse) {
	fmt.Println("Got response!")
	fmt.Println("Name:", fileChunk.Name)
	fmt.Println("Size:", fileChunk.Size)
	fmt.Println("Part:", fileChunk.Part)
	fmt.Println("Total parts:", fileChunk.Parts)
	fmt.Println("MD5:", fileChunk.Md5)
	fmt.Println("Length of Data:", len(fileChunk.Data))
	fmt.Println("Data:", fileChunk.Data)
}
