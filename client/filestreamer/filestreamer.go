// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>

package filestreamer

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	pb "github.com/adirt/rfsb/protos"
	"hash"
	"os"
	"path"
)

var (
	workdir, _          = os.Getwd()
	dirPermissionLevel  = os.FileMode(0755)
	filePermissionLevel = os.FileMode(0644)
)

// The FileStreamer is responsible for writing a single file to disk in chunks
// being streamed into it via Send() and Receive().
// When a FileStreamer's task is over, a call to Close() will close the underlying
// channel and file and returned a calculated MD5 digest of the file.
// Go Pointers: Why I Use Interfaces (in Go), by Kent Rancourt
// medium.com/@kent.rancourt/go-pointers-why-i-use-interfaces-in-go-338ae0bdc9e4
type FileStreamer interface {
	// Send takes a file chunk protobuf as input and writes it to the channel.
	// Send(*pb.FetchResponse)

	// Receive waits to read a file chunk from the channel and should not be called
	// unless a separate goroutine is sending file chunks via Write.
	// Receive() *pb.FetchResponse

	// Write implements the io.Writer interface and writes a file chunk to disk,
	// accumulates the data in a hash and returns the number of bytes that were written.
	Write([]byte) (int, error)

	// Close closes the underlying channel used in Send and Receive and the
	// file pointer used by Write. It returns an MD5 textual representation of
	// the calculated hash.
	// It is recommended to defer Close right after this object as initialized.
	Close() string
}

type fileStreamer struct {
	file          *os.File
	streamChannel chan *pb.FetchResponse
	hash          hash.Hash
}

func NewFileStreamer(filename string) (FileStreamer, error) {
	filePath := path.Join(workdir, filename)
	if err := os.MkdirAll(path.Dir(filePath), dirPermissionLevel); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create directories for file '%s': %s", filePath, err.Error()))
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermissionLevel)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to open file '%s': %s", filePath, err.Error()))
	}

	return &fileStreamer{file: file, streamChannel: make(chan *pb.FetchResponse), hash: md5.New()}, nil
}

func (this *fileStreamer) Send(fileChunk *pb.FetchResponse) {
	this.streamChannel <- fileChunk
}

func (this *fileStreamer) Receive() *pb.FetchResponse {
	return <-this.streamChannel
}

func (this *fileStreamer) Write(fileChunk []byte) (int, error) {
	this.hash.Write(fileChunk)
	return this.file.Write(fileChunk)
}

func (this *fileStreamer) Close() string {
	_ = this.file.Close()
	close(this.streamChannel)
	return hex.EncodeToString(this.hash.Sum(nil))
}
