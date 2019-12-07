/*
Copyright 2019 The Hashmess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "github.com/hasheddan/hashmess/comments/genproto"
)

const (
	address     = "localhost:5050"
	defaultHash = "7dceb8d728e363508013e6698ba0ecf4"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("could not connect to comments service: %v", err)
	}
	defer conn.Close()
	c := pb.NewCommentsServiceClient(conn)

	// Test GetComments()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetComments(ctx, &pb.CommentsRequest{Hash: defaultHash})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Comments: %s", r.GetComments())
}
