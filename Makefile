proto-gen:
	@protoc --go_out=protobuf/auth --go-grpc_out=protobuf/auth \
	        proto/auth.proto
	@protoc --go_out=protobuf/user --go-grpc_out=protobuf/user \
	        proto/user.proto