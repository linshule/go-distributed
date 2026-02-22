package main

import (
	"context"
	"fmt"
	stlog "log"

	"github.com/linshule/go-distributed/provider"
	"github.com/linshule/go-distributed/registry"
	"github.com/linshule/go-distributed/service"
)

func main() {
	host, port := "localhost", "5001"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName: "ProviderService",
		ServiceUrl:  serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, provider.RegisterHandlers)
	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done()

	fmt.Println("Shutting down provider service")
}
