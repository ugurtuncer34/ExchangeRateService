package main

import (
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"

	"exchangerateservice/internal/rates"

	mygrpc "exchangerateservice/internal/grpc" // special name to prevent name conflict
	"exchangerateservice/internal/grpc/pb"
)

func main(){
	// wake up common cache system and background worker
	cache := rates.NewRateCache()
	rates.StartProactiveCache(cache)
	cryptoCache := rates.NewCryptoCache()
	
	// gRPC SERVER PORT: 50051
	// listen to port on TCP
	grpcPort := ":50051"
	grpcListener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	grpcServer := grpc.NewServer() // create gRPC server object
	// create GrpcServer and give cache inside
	myGrpcService := &mygrpc.GrpcServer{
		Cache: cache,
		CryptoCache: cryptoCache,
	}
	// register object to server
	pb.RegisterExchangeRateServiceServer(grpcServer, myGrpcService)

	// start gRPC server inside a Goroutine so that the code can continue reading downward
	go func() {
		log.Printf("gRPC Server is running on port %s", grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// REST API SERVER PORT: 8080
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// create handler and inject cache dependency
	rateHandler := rates.RateHandler{
		Cache: cache,
	}

	// define endpoint
	r.Get("/api/rates", rateHandler.GetRate)

	// start server
	restPort := ":8080"
	log.Printf("Exchange Rate Service API is running on port %s", restPort)

	// http.ListenAndServe blocks the code, listens until close
	err = http.ListenAndServe(restPort, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}