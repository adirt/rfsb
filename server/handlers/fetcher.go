// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	pb "github.com/adirt/rfsb/protos"
	"log"
	"os"
	"path"
	"time"
)

const (
	chunkSize = 64 * 1024
)

type chunkBuffer [chunkSize]byte

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

func (f *Fetcher) HandleRequest(request *pb.FetchRequest, streamChannel chan *pb.FetchResponse) { //([]*pb.FetchResponse, error) {
	responses := make([]*pb.FetchResponse, 0, len(request.Filenames))
	for idx, filename := range request.Filenames {
		log.Printf("Reading file #%d: %s", idx + 1, filename)
		if exists, err := f.pathExists(filename, FILE); !exists {
			if err != nil {
				log.Printf("can't process '%s': %s", filename, err.Error())
			} else {
				log.Printf("file not found: '%s'", filename)
			}
		}
		fileChunkResponses, err := f.readFile(path.Join(f.rootDir, filename), streamChannel)
		if false {
			if err != nil {
				log.Printf("failed to read file '%s': %s", filename, err.Error())
			}
			responses = append(responses, fileChunkResponses...)
		}
	}
	// return responses, nil
}

func (f *Fetcher) readFile(filename string, streamChannel chan *pb.FetchResponse) (fileChunkResponses []*pb.FetchResponse, err error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	fileSize := uint64(fileInfo.Size())
	chunkCount := calculateChunkCount(fileSize)
	hash := md5.New()
	var buffer chunkBuffer
	for chunkIdx := uint64(1); chunkIdx <= chunkCount; chunkIdx++ {
		bytesRead, err := file.Read(buffer[:chunkSize])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to read chunk %d of '%s': %s", chunkIdx, filename, err.Error()))
		}
		hash.Write(buffer[:bytesRead])
		fileChunkResponse := &pb.FetchResponse{
			Name: filename,
			Size: fileSize,
			Data: buffer[:bytesRead],
			Part: chunkIdx,
			Parts: chunkCount,
		}
		if chunkIdx == chunkCount {
			fileChunkResponse.Md5 = hex.EncodeToString(hash.Sum(nil))
		}
		streamChannel <- fileChunkResponse
		time.Sleep(1 * time.Second)
	}
	close(streamChannel)
	return
}

func calculateChunkCount(fileSize uint64) (chunkCount uint64) {
	chunkCount = fileSize / chunkSize
	if fileSize % chunkSize > 0 {
		chunkCount++
	}
	return
}
