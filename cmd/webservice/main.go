package main

import (
	"context"
	"fmt"
	stlog "log"

	"github.com/linshule/go-distributed/registry"
	"github.com/linshule/go-distributed/service"
	"github.com/linshule/go-distributed/web"
)

func main() {
	host, port := "localhost", "5002"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName:    registry.WebService,
		ServiceUrl:     serviceAddress,
		ServiceVersion: "1.0.0",
		Metadata: map[string]string{
			"description": "Web management interface",
			"uiPath":      "/web",
		},
		Tags:           []string{"ui", "management"},
		HealthCheckURL: serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, web.RegisterHandlers)
	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done()

	fmt.Println("Shutting down web service")
}
