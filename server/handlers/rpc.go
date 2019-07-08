// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	pathpkg "path"
)

type resourceType int

const (
	DIR resourceType = iota
	FILE
)

type rpcHandler struct {
	rootDir string
}

func createRpcHandler() (*rpcHandler, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, errors.New("failed to get home directory: " + err.Error())
	}

	rh := &rpcHandler{}
	if exists, err := rh.pathExists(currentUser.HomeDir, DIR); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to validate root dir '%s'", currentUser.HomeDir))
	} else if !exists {
		return nil, errors.New(fmt.Sprintf("root dir '%s' doesn't exist", currentUser.HomeDir))
	}

	rh.rootDir = currentUser.HomeDir
	return rh, nil
}

func (rh rpcHandler) pathExists(path string, resourceType resourceType) (bool, error) {
	absolutePath := pathpkg.Join(rh.rootDir, path)
	stat, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if (resourceType == DIR && !stat.IsDir()) ||
		(resourceType == FILE && stat.IsDir()) {
		return false, nil
	}

	return true, nil
}
