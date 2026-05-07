module vda-e2e

go 1.24.0

require (
	google.golang.org/grpc v1.70.0
	vda v0.0.0
)

replace vda => ../../

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250124145028-65684f501c47 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)
