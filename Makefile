deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
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


# согласно докеру
CLICKHOUSE_DB=default
CLICKHOUSE_USER=root
CLICKHOUSE_PASSWORD="root"

migration-up:
	@docker exec -i clickhouse clickhouse-client --user $(CLICKHOUSE_USER) --password $(CLICKHOUSE_PASSWORD) < ./migrations/001_create_actual_layers_by_client_id.up.sql
	@docker exec -i clickhouse clickhouse-client --user $(CLICKHOUSE_USER) --password $(CLICKHOUSE_PASSWORD) < ./migrations/002_create_history_of_layers_by_client_id.up.sql

migration-down:
	# делаем откат в обратном порядке, jfyi
	@docker exec -i clickhouse clickhouse-client --user $(CLICKHOUSE_USER) --password $(CLICKHOUSE_PASSWORD) -q "DROP TABLE IF EXISTS history_of_layers_by_client_id"
	@docker exec -i clickhouse clickhouse-client --user $(CLICKHOUSE_USER) --password $(CLICKHOUSE_PASSWORD) -q "DROP TABLE IF EXISTS actual_layers_by_client_id"
