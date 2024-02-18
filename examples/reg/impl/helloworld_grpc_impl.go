package impl

import (
	"context"
	"fmt"

	"github.com/baowk/dilu-rd/grpc/pb/service"
)

type TempimplementedGreeterServer struct {
	*service.UnimplementedGreeterServer
}

func (s *TempimplementedGreeterServer) SayHello(ctx context.Context, in *service.HelloRequest) (*service.HelloReply, error) {
	fmt.Println("Name:", in.Name)
	return &service.HelloReply{Message: "Hello " + in.Name}, nil
}
