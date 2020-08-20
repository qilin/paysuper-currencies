.PHONY: grpcgen
grpcgen: ## generate protobuf files
	 protoc pkg/grpc/proto/*.proto --go_out=plugins=grpc:.