// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	pb "github.com/adirt/rfsb/protos"
)

type Browser struct{}

func (b Browser) HandleRequest(request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	dirs := []string{"Videos", "Music", "Documents"}
	filenames := []string{"virus.exe", "dreamboat.png", "cheatsheet.pdf"}
	return &pb.BrowseResponse{Dirs: dirs, Filenames: filenames}, nil
}
