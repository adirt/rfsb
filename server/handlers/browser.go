// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	pb "github.com/adirt/rfsb/protos"
	"log"
)

type Browser struct {
	*rpcHandler
}

func NewBrowser() (*Browser, error) {
	rh, err := newRpcHandler()
	if err != nil {
		return nil, err
	}
	return &Browser{rpcHandler: rh}, nil
}

func (this *Browser) HandleRequest(request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	subdirs, filenames, err := this.dirContents(request.Dir)
	if err != nil {
		log.Printf(err.Error())
		return nil, err
	}
	return &pb.BrowseResponse{Dirs: subdirs, Filenames: filenames}, nil
}
