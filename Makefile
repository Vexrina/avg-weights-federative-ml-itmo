deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go get -u github.com/go-jet/jet/v2
	go install github.com/go-jet/jet/v2/cmd/jet@latest
	go install go.uber.org/mock/mockgen@latest
	@echo "Dependencies installed"


proto: deps
	mkdir -p pkg/avg_weights
	protoc --go_out=pkg --go-grpc_out=pkg \
		api/serverside.proto
	@echo "Proto files generated"

clean:
	rm -rf bin/
	rm -rf pkg/avg_weights/*.pb.go
	rm -rf pkg/avg_weights/*.swagger.json
	@echo "Clean complete"

jet: deps
	jet -source=postgresql -host=localhost -port=5433 -user=metadata_user -password=metadata_pwd -dbname=metadata_db -schema=public -path=generated/

generate: clean deps proto
	@echo "Full generation completed"

containers:
	docker-compose up -d
