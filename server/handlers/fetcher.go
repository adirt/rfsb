// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	pb "github.com/adirt/rfsb/protos"
)

type Fetcher struct {
	*rpcHandler
}

func CreateFetcher() (*Fetcher, error) {
	fetcher := &Fetcher{}
	var err error
	if fetcher.rpcHandler, err = createRpcHandler(); err != nil {
		return nil, err
	}
	return fetcher, nil
}

func (f Fetcher) HandleRequest(request *pb.FetchRequest) (*pb.FetchResponse, error) {
	return &pb.FetchResponse{}, nil
}
