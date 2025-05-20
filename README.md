# Authentification microservice
### Installation and running 
Run docker containers: ```docker compose up --build -d```
Run the app: ```go run ./cmd/auth-service```

### Testing
Install grpc curl: ```go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest```

route /Register: 
```grpcurl -plaintext   -d '{ "email": "test@example.com", "password": "hunter2" }'   localhost:50051   auth.AuthService/Register```
route /ValidateToken:
```grpcurl -plaintext   -d '{ "jwt":  "test_token"}'   localhost:50051   auth.AuthService/ValidateToken```

To stop and remove running container: ```docker compose down```
To regen proto file: ```protoc   -I=./proto   --go_out=./proto   --go_opt=paths=source_relative   --go-grpc_out=./proto   --go-grpc_opt=paths=source_relative   proto/auth.proto```