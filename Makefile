PROTO_DIR := proto
GO_OUT    := $(PROTO_DIR)

.PHONY: proto
proto:
	protoc \
	  -I=$(PROTO_DIR) \
	  --go_out=$(GO_OUT) --go_opt=paths=source_relative \
	  --go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
	  $(PROTO_DIR)/*.proto
