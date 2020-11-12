package listener

import (
	"context"
	pb "github.com/liatrio/rode-api/proto/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var logger *zap.Logger

func TestListener(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Listener Suite")
}

var _ = BeforeSuite(func() {
	logger, _ = zap.NewDevelopment()
})

type mockRodeClient struct {
}

func (m *mockRodeClient) BatchCreateOccurrences(ctx context.Context, in *pb.BatchCreateOccurrencesRequest, opts ...grpc.CallOption) (*pb.BatchCreateOccurrencesResponse, error) {
	return nil, nil
}
