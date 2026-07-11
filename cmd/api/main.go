package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"

	"exchangerateservice/internal/rates"

	mygrpc "exchangerateservice/internal/grpc" // special name to prevent name conflict
	"exchangerateservice/internal/grpc/pb"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// OPENTELEMETRY Tracer
func initTracer() (*sdktrace.TracerProvider, error) {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(), // no password for local test
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceName("FamilyFinance.ExchangeRate"), // name to appear in Jaeger
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main(){
	// start OTel Tracer and clear on close
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

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

	grpcServer := grpc.NewServer( // create gRPC server object and inject OTel Handler
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	) 
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