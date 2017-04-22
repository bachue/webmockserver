package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	proto "webmockserver/proto"

	goflags "github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
)

type Flags struct {
	BindGRPCHost string `long:"grpc-host" description:"Bind GRPC on Host"`
	BindGRPCPort uint16 `long:"grpc-port" description:"Bind GRPC on Port" required:"true"`
	BindHTTPHost string `long:"http-host" description:"Bind HTTP on Host"`
	BindHTTPPort uint16 `long:"http-port" description:"Bind HTTP on Port" required:"true"`
}

var flags Flags

func main() {
	parser := goflags.NewParser(&flags, goflags.HelpFlag|goflags.PassDoubleDash|goflags.IgnoreUnknown)
	if _, err := parser.ParseArgs(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	server := NewServer()

	grpcServer := grpc.NewServer()
	proto.RegisterWebMockServer(grpcServer, server)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", flags.BindHTTPHost, flags.BindHTTPPort),
		Handler:      server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	gracefulShutdown := false

	if grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", flags.BindGRPCHost, flags.BindGRPCPort)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		go func() {
			fmt.Fprintln(os.Stderr, grpcServer.Serve(grpcListener))
			if !gracefulShutdown {
				os.Exit(1)
			}
		}()
	}

	go func() {
		fmt.Fprintln(os.Stderr, httpServer.ListenAndServe())
		if !gracefulShutdown {
			os.Exit(1)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signal := <-signalChannel
	fmt.Printf("Received signal: %s\n", signal.String())
	gracefulShutdown = true
	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
	return
}
