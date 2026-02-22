package main

import (
	"context"
	"fmt"
	stlog "log"

	"github.com/linshule/go-distributed/monitor"
	"github.com/linshule/go-distributed/registry"
	"github.com/linshule/go-distributed/service"
)

func main() {
	host, port := "localhost", "5003"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:    registry.MonitorService,
		ServiceUrl:     serviceAddress,
		ServiceVersion: "1.0.0",
		Metadata: map[string]string{
			"description": "Service health monitoring",
			"checkInterval": "30s",
		},
		Tags:           []string{"monitoring", "health"},
		HealthCheckURL: serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, monitor.RegisterHandlers)
	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done()

	fmt.Println("Shutting down monitor service")
}
