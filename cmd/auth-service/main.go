package authservice
//            • load config  
// │           • init logger  
// │           • connect to Mongo, Redis, NATS  
// │           • wire up adapters → usecases → handlers  
// │           • start gRPC server