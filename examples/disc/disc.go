package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/baowk/dilu-rd/examples/config"
	"github.com/baowk/dilu-rd/grpc/pb/service"
	"github.com/baowk/dilu-rd/models"
	"github.com/baowk/dilu-rd/rd"
)

func main() {
	cfg := config.GetDisConfig()

	rdclient, err := rd.NewRDClient(cfg)
	if err != nil {
		panic(err)
	}
	for {
		for i := 0; i < len(cfg.Discoveries); i++ {
			rs, err := rdclient.GetService(cfg.Discoveries[i].Name, "")
			if err != nil {
				slog.Error("GetService", "err", err)
				time.Sleep(time.Duration(3 * time.Second))
				continue
			}
			if rs != nil {
				slog.Info("service", "name", *rs)
				if rs.Protocol == "http" {
					httpPing(rs)
				} else {
					grpcSayHello(rs)
				}
			} else {
				slog.Error("no service", "name", cfg.Discoveries[0].Name)
			}
		}
		time.Sleep(time.Duration(800 * time.Millisecond))
	}
}

func httpPing(rs *models.ServiceNode) {
	url := rs.GetUrl() + "/ping"
	resp, err := http.Get(url)
	if err != nil {
		rs.IncrFailCnt()
		slog.Error("ping err", "err", err)
		return
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("readall", "err", err)
		return
	}
	slog.Info(string(b))
}

func grpcSayHello(rs *models.ServiceNode) {
	conn, err := rs.GetGrpcConn()
	if err != nil {
		slog.Error("get conn", "err", err)
		return
	}
	c := service.NewGreeterClient(conn)

	r, err := c.SayHello(context.Background(), &service.HelloRequest{Name: "walker"})
	if err != nil {
		slog.Error("could not greet", "err", err)
		rs.IncrFailCnt()
		return
	}
	slog.Info("Greeting: ", "msg", r.Message)
}
