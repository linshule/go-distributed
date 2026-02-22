package main

import (
	"context"
	"fmt"
	stlog "log"

	"github.com/linshule/go-distributed/log"
	"github.com/linshule/go-distributed/registry"
	"github.com/linshule/go-distributed/service"
)

func main() {
	log.Run("./distributed.log")
	host, port := "localhost", "4000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:    registry.LogService,
		ServiceUrl:     serviceAddress,
		ServiceVersion: "1.0.0",
		Metadata: map[string]string{
			"description": "Centralized logging service",
			"logFile":     "./distributed.log",
		},
		Tags:           []string{"logging", "core"},
		HealthCheckURL: serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, log.RegisterHandlers)
	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done()

	fmt.Println("Shutting down log service")
}
