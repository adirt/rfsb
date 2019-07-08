// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"errors"
	"fmt"
	pb "github.com/adirt/rfsb/protos"
	"io/ioutil"
	"path"
)

type Browser struct {
	*rpcHandler
}

func CreateBrowser() (*Browser, error) {
	browser := &Browser{}
	var err error
	if browser.rpcHandler, err = createRpcHandler(); err != nil {
		return nil, err
	}
	return browser, nil
}

func (b Browser) HandleRequest(request *pb.BrowseRequest) (*pb.BrowseResponse, error) {
	if dirExists, err := b.pathExists(request.Dir, DIR); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to validate request dir '%s': %v", request.Dir, err))
	} else if !dirExists {
		return nil, errors.New(fmt.Sprintf("request dir '%s' doesn't exist", request.Dir))
	}

	dirList, err := ioutil.ReadDir(path.Join(b.rootDir, request.Dir))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read request dir '%s': %v", request.Dir, err))
	}

	dirs := make([]string, 0, len(dirList))
	filenames := make([]string, 0, len(dirList))
	for _, item := range dirList {
		if item.IsDir() {
			dirs = append(dirs, item.Name())
		} else {
			filenames = append(filenames, item.Name())
		}
	}

	return &pb.BrowseResponse{Dirs: dirs, Filenames: filenames}, nil
}
