// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	pb "github.com/adirt/rfsb/protos"
)

type Fetcher struct{}

func (f Fetcher) HandleRequest(request *pb.FetchRequest) (*pb.FetchResponse, error) {
	return &pb.FetchResponse{}, nil
}
