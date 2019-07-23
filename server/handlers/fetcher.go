// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package handlers

import (
	"crypto/md5"
	"encoding/hex"
	pb "github.com/adirt/rfsb/protos"
	"log"
	"os"
)

const (
	chunkSize = 64 * 1024
)

type chunkBuffer [chunkSize]byte

type Fetcher struct {
	*rpcHandler
	streamChannel chan *pb.FetchResponse
	fileCountChannel chan int
}

func NewFetcher() (*Fetcher, error) {
	rh, err := newRpcHandler()
	if err != nil {
		return nil, err
	}
	return &Fetcher{
		rpcHandler: rh,
		streamChannel: make(chan *pb.FetchResponse),
		fileCountChannel: make(chan int),
	}, nil
}

func (this *Fetcher) HandleRequest(request *pb.FetchRequest) {
	this.fileCountChannel <- len(request.Filenames)
	this.recursiveStreamDirs(request.Dirs)
	this.streamFiles(request.Filenames)
}

func (this *Fetcher) recursiveStreamDirs(dirs []string) {
	for _, dir := range dirs {
		go func(dir string) {
			log.Printf("Reading directory '%s'", dir)
			subdirs, filenames, err := this.dirContents(dir)
			if err != nil {
                log.Printf(err.Error())
                return
			}
			this.fileCountChannel <- len(filenames)
			this.recursiveStreamDirs(subdirs)
			this.streamFiles(filenames)
		}(dir)
	}
}

func (this *Fetcher) streamFiles(filenames []string) {
	for _, filename := range filenames {
		go func(filename string) {
			log.Printf("Reading file '%s'", filename)
			if exists, err := this.pathExists(filename, FILE); !exists {
				this.fileCountChannel <- -1
				if err != nil {
					log.Printf("can't process '%s': %s", filename, err.Error())
					return
					// TODO: Proper logging and error handling everywhere
				} else {
					log.Printf("file not found: '%s'", filename)
					return
				}
			}

			this.streamFile(filename)
			this.fileCountChannel <- -1
		}(filename)
	}
}

func (this *Fetcher) streamFile(filename string) {
	file, fileSize, err := this.prepareFile(filename)
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
		fileChunkResponse := pb.FetchResponse{
			Name:  filename,
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

func (this *Fetcher) prepareFile(filename string) (file *os.File, fileSize uint64, err error) {
	filePath := this.absolutePath(filename)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}
	fileSize = uint64(fileInfo.Size())
	file, err = os.Open(filePath)
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

func (this *Fetcher) FileCountChannel() chan int {
	return this.fileCountChannel
}

func (this *Fetcher) CloseChannels() {
	close(this.streamChannel)
	close(this.fileCountChannel)
}