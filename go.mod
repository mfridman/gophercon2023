module github.com/mfridman/gophercon2023

go 1.21.1

require (
	connectrpc.com/connect v1.11.1
	github.com/go-chi/chi/v5 v5.0.10
	google.golang.org/protobuf v1.31.0
)

require (
	buf.build/gen/go/mfridman/gophercon2023/connectrpc/go v1.11.1-20230925172220-21748fd4e1ba.1
	buf.build/gen/go/mfridman/gophercon2023/protocolbuffers/go v1.31.0-20230925172220-21748fd4e1ba.1
	github.com/rs/cors v1.10.0
)

require buf.build/gen/go/mfridman/common/protocolbuffers/go v1.31.0-20230923172951-ba92dff3fe5f.1 // indirect
