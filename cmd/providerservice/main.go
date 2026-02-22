package main

import (
	"context"
	"fmt"
	stlog "log"
	"time"

	"github.com/linshule/go-distributed/discovery"
	"github.com/linshule/go-distributed/provider"
	"github.com/linshule/go-distributed/registry"
	"github.com/linshule/go-distributed/service"
)

func main() {
	host, port := "localhost", "5001"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:    registry.ProviderService,
		ServiceUrl:     serviceAddress,
		ServiceVersion: "1.0.0",
		Metadata: map[string]string{
			"description": "Service Provider with Discovery",
		},
		Tags: []string{"discovery", "provider"},
		HealthCheckURL: serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, func() {
		provider.RegisterHandlers()
		discovery.RegisterHandlers()
	})
	if err != nil {
		stlog.Fatalln(err)
	}

	// 启动服务发现定时刷新
	discovery.StartPolling(10 * time.Second)

	<-ctx.Done()

	fmt.Println("Shutting down provider service")
}
