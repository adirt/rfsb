// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
)

type resourceType int

const (
	DIR resourceType = iota
	FILE
)

type rpcHandler struct {
	rootDir string
}

func newRpcHandler() (*rpcHandler, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, errors.New("failed to get home directory: " + err.Error())
	}

	rh := rpcHandler{}
	if exists, err := rh.pathExists(currentUser.HomeDir, DIR); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to validate root dir '%s'", currentUser.HomeDir))
	} else if !exists {
		return nil, errors.New(fmt.Sprintf("root dir '%s' doesn't exist", currentUser.HomeDir))
	}

	rh.rootDir = currentUser.HomeDir
	return &rh, nil
}

func (this *rpcHandler) pathExists(item string, resourceType resourceType) (bool, error) {
	absolutePath := this.absolutePath(item)
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

func (this *rpcHandler) dirContents(dir string) (subdirs []string, filenames []string, err error) {
	exists, err := this.pathExists(dir, DIR)
	if !exists {
		if err != nil {
			err = errors.New(fmt.Sprintf("failed to read dir '%s': %s", dir, err.Error()))
			return
		} else {
			err = errors.New(fmt.Sprintf("dir not found: '%s'", dir))
			return
		}
	}
	dirList, err := ioutil.ReadDir(this.absolutePath(dir))
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to list contents of '%s': %s", dir, err.Error()))
		return
	}

	for _, item := range dirList {
		if item.IsDir() {
			subdirs = append(subdirs, item.Name())
		} else {
			filenames = append(filenames, item.Name())
		}
	}
	return subdirs, filenames, nil
}

func (this *rpcHandler) absolutePath(item string) string {
	if strings.HasPrefix(item, this.rootDir) {
		return item
	}
	return path.Join(this.rootDir, item)
}
