# Authentification microservice
### Installation and running 
Run docker containers: ```docker compose up --build -d```
Run the app: ```go run ./cmd/auth-service```

### Testing
Install grpc curl: ```go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest```

Register an admin = 1: 
```grpcurl -plaintext   -d '{ "email": "test@example.com", "password": "hunter2", "role": 1 }'   localhost:50051   auth.AuthService/Register```

Login:
```grpcurl -plaintext   -d '{ "email": "test@example.com", "password": "hunter2"}'   localhost:50051   auth.AuthService/Login```

Get user by using token:
```grpcurl -plaintext   -d '{ "jwt":  "test_token"}'   localhost:50051   auth.AuthService/ValidateToken```

To stop and remove running container: ```docker compose down```

To regen proto file: ```protoc   -I=./proto   --go_out=./proto   --go_opt=paths=source_relative   --go-grpc_out=./proto   --go-grpc_opt=paths=source_relative   proto/auth.proto```


grpcurl -plaintext   -H "authorization: Bearer <token>"   -d '{ "email": "assan.ayauzhan@gmail.com" }'   localhost:50051   auth.AuthService/ResetPassword

grpcurl -plaintext   -H "authorization: Bearer <token>"   -d '{ "email": "assan.ayauzhan@gmail.com" , "code": "090886", "new_password": "000"}'   localhost:50051   auth.AuthService/ConfirmResetPasswordA