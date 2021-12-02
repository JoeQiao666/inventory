```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=./protobuf  --proto_path=. service.proto
```
```
protoc --go_out=./persistence --go_opt=paths=source_relative --proto_path=./persistence persistence/domain.proto
```