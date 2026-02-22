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
		ServiceName: "WebService",
		ServiceUrl:  serviceAddress,
	}
	ctx, err := service.Start(context.Background(), host, port, r, web.RegisterHandlers)
	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done()

	fmt.Println("Shutting down web service")
}
