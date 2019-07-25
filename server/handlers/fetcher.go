// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"crypto/md5"
	"encoding/hex"
	pb "github.com/adirt/rfsb/protos"
	"log"
	"os"
	"path"
	"sync"
)

const (
	chunkSize = 64 * 1024
)

type chunkBuffer [chunkSize]byte

type Fetcher struct {
	*rpcHandler
	streamChannel chan *pb.FetchResponse
	waitGroup     sync.WaitGroup
}

func NewFetcher() (*Fetcher, error) {
	rh, err := newRpcHandler()
	if err != nil {
		return nil, err
	}
	return &Fetcher{
		rpcHandler:    rh,
		streamChannel: make(chan *pb.FetchResponse),
		waitGroup:     sync.WaitGroup{},
	}, nil
}

func (this *Fetcher) HandleRequest(request *pb.FetchRequest) {
	this.waitGroup.Add(len(request.Filenames))
	this.recursiveStreamDirs(request.Dirs, "")
	this.streamFiles(request.Filenames, "")
	this.waitGroup.Wait()
	close(this.streamChannel)
}

func (this *Fetcher) recursiveStreamDirs(dirs []string, currentDir string) {
	for _, dir := range dirs {
		log.Printf("Reading directory '%s'", dir)
		dirPath := path.Join(currentDir, dir)
		subdirs, filenames, err := this.dirContents(dirPath)
		if err != nil {
			log.Printf(err.Error())
			return
		}
		this.waitGroup.Add(len(filenames))
		this.recursiveStreamDirs(subdirs, dirPath)
		this.streamFiles(filenames, dirPath)
	}
}

func (this *Fetcher) streamFiles(filenames []string, dirPath string) {
	for _, filename := range filenames {
		log.Printf("Reading file '%s'", filename)
		filePath := path.Join(dirPath, filename)
		go this.streamFile(filePath)
	}
}

func (this *Fetcher) streamFile(filePath string) {
	defer this.waitGroup.Done()
	file, fileSize, err := this.prepareFile(filePath)
	if err != nil {
		log.Printf("failed to open file '%s': %s", filePath, err.Error())
		return
	}
	defer file.Close()

	chunkCount := calculateChunkCount(fileSize)
	hash := md5.New()
	for chunkIdx := uint64(1); chunkIdx <= chunkCount; chunkIdx++ {
		var buffer chunkBuffer
		bytesRead, err := file.Read(buffer[:])
		if err != nil {
			log.Printf("failed to read chunk %d of '%s': %s", chunkIdx, filePath, err.Error())
			return
		}

		hash.Write(buffer[:bytesRead])
		fileChunkResponse := pb.FetchResponse{
			Name:  filePath,
			Size:  fileSize,
			Data:  buffer[:bytesRead],
			Part:  chunkIdx,
			Parts: chunkCount,
		}
		if chunkIdx == chunkCount {
			fileChunkResponse.Md5 = hex.EncodeToString(hash.Sum(nil))
		}

		this.streamChannel <- &fileChunkResponse
	}
}

func (this *Fetcher) prepareFile(filePath string) (file *os.File, fileSize uint64, err error) {
	fileAbsPath := this.absolutePath(filePath)
	fileInfo, err := os.Stat(fileAbsPath)
	if err != nil {
		return
	}
	fileSize = uint64(fileInfo.Size())
	file, err = os.Open(fileAbsPath)
	return
}

func calculateChunkCount(fileSize uint64) uint64 {
	chunkCount := fileSize / chunkSize
	if fileSize%chunkSize > 0 {
		chunkCount++
	}
	return chunkCount
}

func (this *Fetcher) StreamChannel() chan *pb.FetchResponse {
	return this.streamChannel
}
