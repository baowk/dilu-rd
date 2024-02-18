package main

import (
	"context"
	"fmt"
	"io"
	"log"
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
		if len(cfg.Discoveries) > 0 {
			for i := 0; i < len(cfg.Discoveries); i++ {
				rs, err := rdclient.GetService(cfg.Discoveries[i].Name, "")
				if err != nil {
					fmt.Println(err)
					time.Sleep(time.Duration(3 * time.Second))
					continue
				}
				if rs != nil {
					logger.Info("service", zap.Any("name", *rs))
					if rs.Protocol == "http" {
						Get(rs, logger.Sugar())
					} else {
						fmt.Println("grpc")
						conn, err := rs.GetGrpcConn()
						if err != nil {
							logger.Error("get conn", zap.Error(err))
							return
						}
						c := service.NewGreeterClient(conn)

						r, err := c.SayHello(context.Background(), &service.HelloRequest{Name: "walker"})
						if err != nil {
							log.Println("could not greet", err)
							return
						}
						log.Println("Greeting: ", r.Message)
					}
				} else {
					logger.Error("no service", zap.Any("name", cfg.Discoveries[0].Name))
				}
			}
		}
		time.Sleep(time.Duration(800 * time.Millisecond))
	}
}

func Get(rs *models.ServiceNode, logger *zap.SugaredLogger) {
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
	fmt.Println(string(b))
}
