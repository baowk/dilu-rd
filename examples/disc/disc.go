package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/baowk/dilu-rd/examples/config"
	"github.com/baowk/dilu-rd/grpc/pb/service"
	"github.com/baowk/dilu-rd/models"
	"github.com/baowk/dilu-rd/rd"

	"go.uber.org/zap"
)

func main() {
	cfg := config.GetDisConfig()
	logger, _ := zap.NewDevelopment()

	rdclient, err := rd.NewRDClient(cfg, logger.Sugar())
	if err != nil {
		panic(err)
	}
	for {
		for i := 0; i < len(cfg.Discoveries); i++ {
			rs, err := rdclient.GetService(cfg.Discoveries[i].Name, "")
			if err != nil {
				logger.Error("GetService", zap.Error(err))
				time.Sleep(time.Duration(3 * time.Second))
				continue
			}
			if rs != nil {
				logger.Info("service", zap.Any("name", *rs))
				if rs.Protocol == "http" {
					httpPing(rs, logger.Sugar())
				} else {
					grpcSayHello(rs, logger.Sugar())
				}
			} else {
				logger.Error("no service", zap.Any("name", cfg.Discoveries[0].Name))
			}
		}
		time.Sleep(time.Duration(800 * time.Millisecond))
	}
}

func httpPing(rs *models.ServiceNode, logger *zap.SugaredLogger) {
	url := rs.GetUrl() + "/ping"
	resp, err := http.Get(url)
	if err != nil {
		rs.IncrFailCnt()
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(b))
}

func grpcSayHello(rs *models.ServiceNode, logger *zap.SugaredLogger) {
	conn, err := rs.GetGrpcConn()
	if err != nil {
		logger.Error("get conn", zap.Error(err))
		return
	}
	c := service.NewGreeterClient(conn)

	r, err := c.SayHello(context.Background(), &service.HelloRequest{Name: "walker"})
	if err != nil {
		logger.Error("could not greet", zap.Error(err))
		return
	}
	logger.Info("Greeting: ", zap.String("msg", r.Message))
}
