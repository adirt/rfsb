// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/adirt/rfsb/client/filestreamer"
	pb "github.com/adirt/rfsb/protos"
	"google.golang.org/grpc"
	"log"
	"strings"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := pb.NewRemoteFileSystemBrowserClient(conn)

	// testBrowse(client)
	testFetch(client)
}

func testBrowse(client pb.RemoteFileSystemBrowserClient) {
	timedCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.Browse(timedCtx, &pb.BrowseRequest{Dir: "Projects/rust/sopcast/target/debug"})
	if err != nil {
		log.Fatalf("failed to get browse response: %v", err)
	}
	fmt.Printf("Dirs: %s\nFilenames: %s\n",
		strings.Join(response.Dirs, ", "),
		strings.Join(response.Filenames, ", "))
}

func testFetch(client pb.RemoteFileSystemBrowserClient) {
	fetchRequest := pb.FetchRequest{
		Dirs:      []string{"Games", "bin"},
		Filenames: []string{"Pictures/Screenshot from 2019-06-15 18-43-14.png"},
	}
	fetchClient, err := client.Fetch(context.Background(), &fetchRequest)
	if err != nil {
		log.Fatalf("failed to get fetch client: %v", err)
	}

	nameToStreamerMap := map[string]filestreamer.FileStreamer{}
	doneReceiving := false

	for !doneReceiving {
		response, err := fetchClient.Recv()
		if err != nil {
			log.Printf("failed to get fetch response: %v", err)
			break
		}
		printChunkInfo(response)

		fileStreamer, exists := nameToStreamerMap[response.Name]
		if !exists {
			fileStreamer, err = filestreamer.NewFileStreamer(response.Name)
			if err != nil {
				log.Printf("failed to initialize a file streamer for '%s': %s", response.Name, err.Error())
				continue
			}
			nameToStreamerMap[response.Name] = fileStreamer
		}

		if _, err = fileStreamer.Write(response.Data); err != nil {
			log.Printf("failed to write %d bytes of part (%d/%d) of '%s' to disk: %s",
				len(response.Data), response.Part, response.Parts, response.Name, err.Error())
		}

		doneReceiving = finalizeChunk(response, nameToStreamerMap)
	}
}

func finalizeChunk(response *pb.FetchResponse, nameToStreamerMap map[string]filestreamer.FileStreamer) (doneReceiving bool) {
	if response.Part < response.Parts {
		return
	}

	md5digest := nameToStreamerMap[response.Name].Close()
	fmt.Printf("Server MD5: %s\nClient MD5: %s\n", response.Md5, md5digest)
	if response.Md5 == md5digest {
		fmt.Println("Hooray, both MD5's match!!!")
	} else {
		fmt.Println("Damn, the MD5's don't match...")
	}

	delete(nameToStreamerMap, response.Name)
	if len(nameToStreamerMap) == 0 {
		doneReceiving = true
	}
	return
}

func printChunkInfo(fileChunk *pb.FetchResponse) {
	// for debugging purposes only
	fmt.Println("====================================================")
	fmt.Println("Name:", fileChunk.Name)
	fmt.Println("Size:", fileChunk.Size)
	fmt.Println("Part:", fileChunk.Part)
	fmt.Println("Total parts:", fileChunk.Parts)
	fmt.Println("MD5:", fileChunk.Md5)
	fmt.Println("Length of Data:", len(fileChunk.Data))
	fmt.Println("====================================================")
	// fmt.Println("Data:", fileChunk.Data)
}

func testFetchOld(client pb.RemoteFileSystemBrowserClient) {
	fetchClient, err := client.Fetch(context.Background(), &pb.FetchRequest{Filenames: []string{"Games/Mednafen/Genesis/Road Rash 3 (UEJ) [!].zip"}})
	if err != nil {
		log.Fatalf("failed to get fetch client: %v", err)
	}
	var incomingData []byte
	hash := md5.New()
	for {
		response, err := fetchClient.Recv()
		if err != nil {
			log.Fatalf("failed to get fetch response: %v", err)
		}
		printChunkInfo(response)
		incomingData = append(incomingData, response.Data...)
		hash.Write(response.Data)
		if response.Part == response.Parts {
			clientCalculatedMd5 := hex.EncodeToString(hash.Sum(nil))
			fmt.Printf("Server MD5: %s\nClient MD5: %s\n", response.Md5, clientCalculatedMd5)
			if response.Md5 == clientCalculatedMd5 {
				fmt.Println("Hooray, both MD5's match!!!")
			} else {
				fmt.Println("Damn, the MD5's don't match...")
			}
			return
		}
	}
}
