// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"crypto/md5"
	"encoding/hex"
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

func (f *Fetcher) HandleRequest(request *pb.FetchRequest, streamChannel chan *pb.FetchResponse) {
	for idx, filename := range request.Filenames {
		go func(idx int, filename string) {
			log.Printf("Reading file #%d: %s", idx+1, filename)
			if exists, err := f.pathExists(filename, FILE); !exists {
				if err != nil {
					log.Printf("can't process '%s': %s", filename, err.Error())
					return
					// TODO: Proper logging and error handling everywhere
				} else {
					log.Printf("file not found: '%s'", filename)
					return
				}
			}
			streamFile(path.Join(f.rootDir, filename), streamChannel)
		}(idx, filename)

		time.Sleep(5 * time.Second) // TODO: learn how to use goroutines with channels for real
		close(streamChannel)
	}
}

func streamFile(filename string, streamChannel chan *pb.FetchResponse) {
	file, fileSize, err := prepareFile(filename)
	if err != nil {
		log.Printf("failed to open file '%s': %s", filename, err.Error())
		return
	}
	defer file.Close()

	chunkCount := calculateChunkCount(fileSize)
	hash := md5.New()
	for chunkIdx := uint64(1); chunkIdx <= chunkCount; chunkIdx++ {
		var buffer chunkBuffer
		bytesRead, err := file.Read(buffer[:])
		if err != nil {
			log.Printf("failed to read chunk %d of '%s': %s", chunkIdx, filename, err.Error())
			return
		}

		hash.Write(buffer[:bytesRead])
		fileChunkResponse := &pb.FetchResponse{
			Name:  filename,
			Size:  fileSize,
			Data:  buffer[:bytesRead],
			Part:  chunkIdx,
			Parts: chunkCount,
		}
		if chunkIdx == chunkCount {
			fileChunkResponse.Md5 = hex.EncodeToString(hash.Sum(nil))
		}

		streamChannel <- fileChunkResponse
	}
}

func prepareFile(filename string) (file *os.File, fileSize uint64, err error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return
	}
	fileSize = uint64(fileInfo.Size())
	file, err = os.Open(filename)
	return
}

func calculateChunkCount(fileSize uint64) (chunkCount uint64) {
	chunkCount = fileSize / chunkSize
	if fileSize%chunkSize > 0 {
		chunkCount++
	}
	return
}
