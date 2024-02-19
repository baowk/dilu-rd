package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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

	//服务启动参数
	srv := &http.Server{
		Addr:    "0.0.0.0:5000",
		Handler: r,
	}

	//启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen: ", err)
		}
	}()

	//r.Run(":5000")

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("Shutdown Server " + time.Now().String())

	rdclient.Deregister()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
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
