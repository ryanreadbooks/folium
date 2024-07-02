package server

import (
	"context"
	"log"
	"net"

	apiv1 "github.com/ryanreadbooks/folium/api/v1"
	"github.com/ryanreadbooks/folium/internal/pkg"
	"github.com/ryanreadbooks/folium/internal/segment/idgen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	serverGrpc *grpc.Server
)

func InitGrpc() {
	serverGrpc = grpc.NewServer()
	apiv1.RegisterFoliumServiceServer(serverGrpc, &grpcServer{})

	listener, err := net.Listen("tcp", ":9528")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := serverGrpc.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()
}

func CloseGrpc() {
	if serverGrpc != nil {
		serverGrpc.GracefulStop()
	}
}

type grpcServer struct {
	apiv1.UnimplementedFoliumServiceServer
}

func (s *grpcServer) Next(ctx context.Context, req *apiv1.NextRequest) (*apiv1.NextResponse, error) {
	id, err := idgen.GetNext(ctx, req.Key, idgen.WithStep(req.Step))
	if err != nil {
		pkgerr, ok := err.(*pkg.Err)
		if ok {
			return nil, status.Error(codes.Code(pkgerr.Code), pkgerr.Msg)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &apiv1.NextResponse{
		Id: id,
	}, nil
}

func (s *grpcServer) Ping(ctx context.Context, in *apiv1.PingRequest) (*apiv1.PingResponse, error) {
	return &apiv1.PingResponse{}, nil
}
