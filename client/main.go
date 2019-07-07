// Copyright © 2019 Adir Tzuberi <adir85@gmail.com>
package main

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := client.Browse(ctx, &pb.BrowseRequest{Dir: "Go Rules"})
	if err != nil {
		log.Fatalf("failed to get browse response: %v", err)
	}
	fmt.Printf("Dirs: %s\nFilenames: %s\n",
		strings.Join(response.GetDirs(), ", "),
		strings.Join(response.GetFilenames(), ", "))
}
