package main

import (
	"context"
	"fmt"
	stlog "log"

	"github.com/linshule/go-distributed/library"
	"github.com/linshule/go-distributed/registry"
	"github.com/linshule/go-distributed/service"
)

func main() {
	host, port := "localhost", "5000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:    registry.LibraryService,
		ServiceUrl:     serviceAddress,
		ServiceVersion: "1.0.0",
		Metadata: map[string]string{
			"description": "Library management service",
			"dependsOn":   "LogService",
		},
		Tags:           []string{"library", "business"},
		HealthCheckURL: serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, library.RegisterHandlers)
	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done()

	fmt.Println("Shutting down library service")
}
