package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/baowk/dilu-rd/examples/config"
	"github.com/baowk/dilu-rd/examples/reg/impl"
	"github.com/baowk/dilu-rd/grpc/pb/health"
	"github.com/baowk/dilu-rd/grpc/pb/service"
	"github.com/baowk/dilu-rd/rd"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	r := gin.Default()
	r.GET("/api/health", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ping": "pong", "time": time.Now()})
	})

	cfg := config.GetConfig()

	ips := GetLocalHost()

	logger, _ := zap.NewDevelopment()

	rdclient, err := rd.NewRDClient(cfg, logger.Sugar())
	if err != nil {
		logger.Debug("NewRDClient err:", zap.Error(err))
	}

	logger.Debug("rdclient:", zap.Any("client", rdclient))

	go func() { //grpc服务
		lis, err := net.Listen("tcp", ":5001")
		if err != nil {
			logger.Error("failed to listen", zap.Error(err))
		}
		s := grpc.NewServer()
		health.RegisterHealthServer(s, &health.HealthServerImpl{})
		service.RegisterGreeterServer(s, &impl.TempimplementedGreeterServer{})
		fmt.Println("grpc server start", ips, "5001")
		logger.Debug("grpc start:", zap.String("ip", ips), zap.Int("port", 5001))
		err = s.Serve(lis)
		if err != nil {
			logger.Error("failed to serve", zap.Error(err))
		}
	}()

	logger.Debug("http start:", zap.String("ip", ips), zap.Int("port", 5000))

	r.Run(":5000")
}

func GetLocalHost() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}

	}
	return ""
}
