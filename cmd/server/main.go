package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"avg_weights_fed_ml_itmo/internal/app"
	"avg_weights_fed_ml_itmo/internal/app/usecase"
	"avg_weights_fed_ml_itmo/internal/cron/aggregator"
	"avg_weights_fed_ml_itmo/internal/minio_repo"
	"avg_weights_fed_ml_itmo/pkg/serverside"
)

const grpcPort = "9090"

func main() {
	ctx := context.Background()

	minioRepo := minio_repo.NewMinioRepo(
		"localhost:9000",
		"admin",
		"admin12345",
		"mybucket",
	)

	aggr := aggregator.NewAggregator(minioRepo)

	amw := usecase.NewUpserter(minioRepo)

	service := app.NewService(amw)
	server := grpc.NewServer()

	serverside.RegisterAvgWeightsServer(server, service)
	reflection.Register(server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		aggr.Aggregate(ctx)
	}()
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err = server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")
	server.GracefulStop()
}
