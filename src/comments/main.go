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
	"database/sql"
	"fmt"
	"net"
	"os"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	pb "github.com/hasheddan/hashmess/src/comments/genproto"
)

const listenPort = "5050"

var db *sql.DB

type commentsService struct{}

func main() {
	var (
		connectionName = os.Getenv("CLOUDSQL_CONNECTION_NAME")
		user           = os.Getenv("CLOUDSQL_USER")
		password       = os.Getenv("CLOUDSQL_PASSWORD")
		database       = os.Getenv("CLOUDSQL_DATABASE")
	)

	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@cloudsql(%s)/%s", user, password, connectionName, database))
	if err != nil {
		log.Fatalf("Could not open db: %v", err)
	}

	port := listenPort
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	svc := new(commentsService)

	srv := grpc.NewServer()
	log.Infof("starting comments service on tcp: %q", lis.Addr().String())
	pb.RegisterCommentsServiceServer(srv, svc)
	health.RegisterHealthServer(srv, svc)
	err = srv.Serve(lis)
	log.Fatal(err)
}

func (cs *commentsService) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{Status: health.HealthCheckResponse_SERVING}, nil
}

func (cs *commentsService) Watch(in *health.HealthCheckRequest, stream health.Health_WatchServer) error {
	// Create update channel for stream.
	update := make(chan health.HealthCheckResponse_ServingStatus, 1)

	// Write initial status to stream.
	update <- health.HealthCheckResponse_SERVING

	var lastSentStatus health.HealthCheckResponse_ServingStatus = -1
	for {
		select {
		// Read status from channel.
		case servingStatus := <-update:
			// If status has not changed do not send to stream.
			if lastSentStatus == servingStatus {
				continue
			}
			lastSentStatus = servingStatus
			err := stream.Send(&health.HealthCheckResponse{Status: servingStatus})
			if err != nil {
				return status.Error(codes.Canceled, "Stream has ended.")
			}
		// Stop polling if context is done.
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream has ended.")
		}
	}
}

func (cs *commentsService) GetComments(ctx context.Context, req *pb.CommentsRequest) (*pb.Comments, error) {
	rows, err := db.Query("SELECT user, message FROM comments WHERE hash=?", req.GetHash())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*pb.Comment{}

	for rows.Next() {
		var user string
		var message string
		if err := rows.Scan(&user, &message); err != nil {
			return nil, err
		}
		comments = append(comments, &pb.Comment{User: user, Message: message})
	}
	return &pb.Comments{
		Comments: comments,
	}, nil
}
