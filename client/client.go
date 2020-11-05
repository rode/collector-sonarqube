package client

import (
	"context"
	"log"
	"time"
	"fmt"

	// "github.com/grafeas/grafeas/proto/v1beta1/common_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/grafeas_go_proto"
	// "github.com/grafeas/grafeas/proto/v1beta1/package_go_proto"
	// "github.com/grafeas/grafeas/proto/v1beta1/vulnerability_go_proto"

	pb "github.com/liatrio/rode-collector-service/proto/v1alpha1"
	"google.golang.org/grpc"
	// "google.golang.org/protobuf/types/known/timestamppb"
)

// const (
// 	address = "localhost:50051"
// )

type RodeClient struct {
	URL string
}

func (rc *RodeClient) SendOccurrences(occurrences []*grafeas_go_proto.Occurrence) error {
	conn, err := rc.grpcConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	c := pb.NewRodeCollectorServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := c.BatchCreateOccurrences(ctx, &grafeas_go_proto.BatchCreateOccurrencesRequest{
		Occurrences: occurrences,
		Parent:      "projects/test123",
	})
	if err != nil {
		log.Fatalf("could not create occurrence: %v", err)
	}
	fmt.Printf("%#v\n", response)
	return nil
}

func (rc *RodeClient) grpcConnection() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(rc.URL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("error establishing gRPC connection with Rode API: %s", err)
	}

	return conn, nil
}
