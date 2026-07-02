.PHONY: install gen

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest

gen:
	@echo "Creating pb directory..."
	mkdir -p pb
	@echo "Downloading proto dependencies..."
	mkdir -p proto/google/api proto/validate
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o proto/google/api/annotations.proto
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o proto/google/api/http.proto
	curl -sSL https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto -o proto/validate/validate.proto
	@echo "Generating Go code from proto..."
	PATH="$$(go env GOPATH)/bin:$$PATH" protoc -I ./proto \
		--go_out=./pb --go_opt=paths=source_relative \
		--go-grpc_out=./pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=./pb --grpc-gateway_opt=paths=source_relative \
		--validate_out="lang=go,paths=source_relative:./pb" \
		proto/service.proto
	@echo "Code generation successful!"

run:
	go run cmd/server/main.go

watch:
	@echo "Starting Air for live reloading..."
	PATH="$$(go env GOPATH)/bin:$$PATH" air