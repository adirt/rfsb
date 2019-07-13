// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	pb "github.com/adirt/rfsb/protos"
	"google.golang.org/grpc"
	"log"
	"strings"
	"time"
)

func main() {
	const address = "localhost:50051"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := pb.NewRemoteFileSystemBrowserClient(conn)

	// Test Browse
	timedCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.Browse(timedCtx, &pb.BrowseRequest{Dir: "Projects/rust/sopcast/target/debug"})
	if err != nil {
		log.Fatalf("failed to get browse response: %v", err)
	}
	fmt.Printf("Dirs: %s\nFilenames: %s\n",
		strings.Join(response.Dirs, ", "),
		strings.Join(response.Filenames, ", "))

	// Test Fetch
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
		fmt.Println("Got response!")
		fmt.Println("Name:", response.Name)
		fmt.Println("Size:", response.Size)
		fmt.Println("Part:", response.Part)
		fmt.Println("Total parts:", response.Parts)
		fmt.Println("MD5:", response.Md5)
		fmt.Println("Length of Data:", len(response.Data))
		fmt.Println("Data:", response.Data)
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
