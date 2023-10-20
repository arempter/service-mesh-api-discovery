package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"service-mesh-api-discovery/pkg/adapter"
	"service-mesh-api-discovery/pkg/k8s"
	"service-mesh-api-discovery/pkg/storage"
	"syscall"
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

func main() {

	port := flag.Int("port", 65000, "listen port")
	flag.Parse()

	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		slog.Error("failed to listen", "err", err.Error())
		os.Exit(1)
	}

	// k8s collector
	collector := k8s.NewK8sCollector(k8s.Clientset())

	// storage dependency and http rest server
	storage := storage.NewInMemStore()
	go setupHttpServer(storage)

	// register logs recveiver
	accesslogv3.RegisterAccessLogServiceServer(srv, adapter.NewAdapter(collector, storage))

	go func() {
		slog.Info("starting envoy log receiver on", "port", *port)
		err = srv.Serve(lis)
		if err != nil {
			slog.Error("failed to start grpc server", "err", err.Error())
		}
	}()

	stop := make(chan struct{})
	go handleShutdown(stop)

	// run k8s collector
	collector.Run(stop)
}

func handleShutdown(stop chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGABRT, os.Interrupt)

	<-c
	slog.Info("Shutting down")
	close(stop)
	// poors man grace period
	time.Sleep(3 * time.Second)
	os.Exit(0)
}

func setupHttpServer(inMemStore storage.InMemStore) {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/endpoints", apiEndpointHandler(inMemStore))

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		slog.Error("failed to start api endpoints rest server", "err", err.Error())
	}
}

func apiEndpointHandler(inMemStore storage.InMemStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		js, err := json.Marshal(inMemStore.Get())
		if err != nil {
			slog.Error("failed to serialize response", "err", err.Error())
		}
		w.Write(js)
	}
}
