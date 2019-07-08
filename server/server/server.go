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

func (s *rfsbServer) Serve() error {
	grpcServer := grpc.NewServer()
	pb.RegisterRemoteFileSystemBrowserServer(grpcServer, s)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.New("failed to listen: " + err.Error())
	}
	if err := grpcServer.Serve(listener); err != nil {
		return errors.New("failed to serve: " + err.Error())
	}
	return nil
}

func (s *rfsbServer) Browse(ctx context.Context, request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	log.Printf("Received browse request for %s", request.Dir)
	response, err := s.browser.HandleRequest(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *rfsbServer) Fetch(request *pb.FetchRequest, stream pb.RemoteFileSystemBrowser_FetchServer) error {
	log.Printf("Received fetch request")
	response, err := s.fetcher.HandleRequest(request)
	if err != nil {
		return err
	}
	log.Println(response)
	return nil
}
