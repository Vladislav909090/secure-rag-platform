module secure-rag-platform/services/ai-inference

go 1.25.0

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260610212136-7ab31c22f7ad
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
	secure-rag-platform/api v0.0.0
)

replace secure-rag-platform/api => ../../api

require (
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260610212136-7ab31c22f7ad // indirect
)
