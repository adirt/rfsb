// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package server

import (
	"code.cloudfoundry.org/bytefmt"
	"context"
	"errors"
	"fmt"
	pb "github.com/adirt/rfsb/protos"
	"github.com/adirt/rfsb/server/handlers"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

const (
	port = 50051
)

type RfsbServer struct{}

func (this *RfsbServer) Serve() error {
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

func (this *RfsbServer) Browse(ctx context.Context, request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	log.Printf("Received browse request for %s", request.Dir)

	browser, err := handlers.NewBrowser()
	if err != nil {
		log.Fatalf("failed to initialize browse handler: %s", err.Error())
	}

	response, err := browser.HandleRequest(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (this *RfsbServer) Fetch(request *pb.FetchRequest, stream pb.RemoteFileSystemBrowser_FetchServer) error {
	log.Printf("Received fetch request")

	fetcher, err := handlers.NewFetcher()
	if err != nil {
		log.Fatalf("failed to initialize fetch handler: %s", err.Error())
	}

	var ( // fetch stats
		chunkCount  = uint64(0)
		fileCount   = uint64(0)
		totalSize   = uint64(0)
		timeStart   = time.Now()
		timeEnd     time.Time
		timeElapsed time.Duration
	)

	go fetcher.HandleRequest(request)
	for fileChunk := range fetcher.StreamChannel() {
		printChunkInfo(fileChunk)
		if err := stream.Send(fileChunk); err != nil {
			log.Printf("failed to stream requested data at '%s' part %d: %s", fileChunk.Name, fileChunk.Part, err.Error())
			return err
		}

		chunkCount++
		totalSize += uint64(len(fileChunk.Data))
		if fileChunk.Part == fileChunk.Parts {
			fileCount++
		}
	}

	timeEnd = time.Now()
	timeElapsed = timeEnd.Sub(timeStart)
	log.Printf("Transferred %d chunks, %d files, total size %s in %s", chunkCount, fileCount, bytefmt.ByteSize(totalSize), timeElapsed)
	return nil
}

func printChunkInfo(fileChunk *pb.FetchResponse) {
	// for debugging purposes only
	fmt.Println("====================================================")
	fmt.Println("Name:", fileChunk.Name)
	fmt.Println("Size:", fileChunk.Size)
	fmt.Println("Part:", fileChunk.Part)
	fmt.Println("Total parts:", fileChunk.Parts)
	fmt.Println("MD5:", fileChunk.Md5)
	fmt.Println("Length of Data:", len(fileChunk.Data))
	fmt.Println("====================================================")
	// fmt.Println("Data:", fileChunk.Data)
}
