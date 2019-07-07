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
	"os"
)

const (
	port = 50051
)

type rfsbServer struct {
	rootDir string
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

func NewServer(rootDir string) (*rfsbServer, error) {
	server := &rfsbServer{}
	if err := os.Chdir(rootDir); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to switch to root directory at %s: %v\n", rootDir, err))
	}
	server.rootDir = rootDir
	return server, nil
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

func (s *rfsbServer) Browse(context context.Context, request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	log.Printf("Received browse request for %s/%s", s.rootDir, request.Dir)
	browser := handlers.Browser{}
	response, err := browser.HandleRequest(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *rfsbServer) Fetch(request *pb.FetchRequest, stream pb.RemoteFileSystemBrowser_FetchServer) error {
	log.Printf("Received fetch request")
	fetcher := handlers.Fetcher{}
	response, err := fetcher.HandleRequest(request)
	if err != nil {
		return err
	}
	log.Println(response)
	return nil
}
