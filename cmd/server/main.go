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

func main() {
	ctx := context.Background()

	// attention: фоллбеки предназначены для локального запуска, не для контейнерного запуска сервиса!
	var (
		minioEndpoint  = getEnv("MINIO_ENDPOINT", "localhost:9000")
		minioAccessKey = getEnv("MINIO_ACCESS_KEY", "admin")
		minioSecretKey = getEnv("MINIO_SECRET_KEY", "admin12345")
		minioBucket    = getEnv("MINIO_BUCKET", "mybucket")
	)
	minioRepo := minio_repo.NewMinioRepo(
		minioEndpoint,
		minioAccessKey,
		minioSecretKey,
		minioBucket,
	)

	var (
		pathToPythonScript = getEnv("PATH_TO_PYTHON_SCRIPT", "internal/cron/aggregator/")
		pythonBinPath      = getEnv("PYTHON_BIN_PATH", ".venv/bin/python")
	)

	aggr := aggregator.NewAggregator(minioRepo, pathToPythonScript, pythonBinPath)

	amw := usecase.NewUpserter(minioRepo)
	service := app.NewService(amw)
	server := grpc.NewServer()

	serverside.RegisterAvgWeightsServer(server, service)
	reflection.Register(server)

	grpcPort := getEnv("GRPC_PORT", "8081")
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
