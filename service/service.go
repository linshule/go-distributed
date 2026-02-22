package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/linshule/go-distributed/registry"
)

func Start(ctx context.Context, host, port string, reg registry.Registration, registerHandlersFunc func()) (context.Context, error) {
	registerHandlersFunc()
	ctx = startServer(ctx, reg.ServiceName, host, port)
	err := registry.RegistrationService(reg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func startServer(ctx context.Context, serviceName registry.ServiceName, host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = host + ":" + port
	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()
	go func() {
		fmt.Printf("%v started.Press any key to stop.\n", serviceName)
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()
	return ctx
}
