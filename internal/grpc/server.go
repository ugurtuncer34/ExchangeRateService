package grpc

import (
	"context"
	"exchangerateservice/internal/grpc/pb" // package of files generated with Protoc
	"exchangerateservice/internal/rates"   // own business logic (cache, tcmb)
	"fmt"
	"log"
	"time"
)

type GrpcServer struct{ // will apply the methods in the contract
	pb.UnimplementedExchangeRateServiceServer // if new methods, don't crash
	Cache *rates.RateCache // work on the same cache (ram) as in REST
} 

// this func is the corresponding method in rpc for Go
func (s *GrpcServer) GetExchangeRate(ctx context.Context, req *pb.RateRequest) (*pb.RateResponse, error) {
	// no Json parse, direct data as struct
	targetDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		// no json on error, gRPC has error structure
		return nil, fmt.Errorf("invalid date format: %v", err)
	}

	cacheKey := fmt.Sprintf("%s_%s", req.Currency, targetDate.Format("2006-01-02"))
	// cache control
	if cachedRate, exists := s.Cache.Get(cacheKey); exists {
		log.Printf("[gRPC] Served %s from CACHE", cacheKey)

		// RateResponse type created with protobuf
		return &pb.RateResponse{
			Currency: req.Currency,
			Date: req.Date,
			Rate: cachedRate,
			Source: "Cache",
		}, nil
	}

	// if not in cache, fetch from Tcmb
	rate, err := rates.FetchRateByDate(req.Currency, targetDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rate: %v", err)
	}

	// write to cache
	s.Cache.Set(cacheKey, rate)
	log.Printf("[gRPC] Served %s from TCMB", cacheKey)

	// return success response
	return &pb.RateResponse{
		Currency: req.Currency,
		Date: req.Date,
		Rate: rate,
		Source: "TCMB",
	}, nil
}