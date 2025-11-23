deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Dependencies installed"


proto: deps
	mkdir -p pkg/avg_weights
	protoc --go_out=pkg --go-grpc_out=pkg \
		api/serverside.proto
	@echo "Proto files generated"

# Генерация swagger/openapi документации
swagger: proto
	protoc -Iapi \
       --openapiv2_out=pkg/avg_weights \
       --openapiv2_opt logtostderr=true \
       api/serverside.proto

clean:
	rm -rf bin/
	rm -rf pkg/avg_weights/*.pb.go
	rm -rf pkg/avg_weights/*.swagger.json
	@echo "Clean complete"

generate: clean deps proto swagger
	@echo "Full generation completed"


containers:
	docker-compose up -d