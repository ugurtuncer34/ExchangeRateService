package grpc

import (
	"context"
	"exchangerateservice/internal/grpc/pb" // package of files generated with Protoc
	"exchangerateservice/internal/rates"   // own business logic (cache, tcmb)
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/metadata"
)

type GrpcServer struct{ // will apply the methods in the contract
	pb.UnimplementedExchangeRateServiceServer // if new methods, don't crash
	Cache *rates.RateCache // work on the same cache (ram) as in REST
	CryptoCache *rates.CryptoCache
} 

// new logging structure
func logWithCorrelation(ctx context.Context, format string, v ...interface{}) {
	corrID := "unknown"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-correlation-id"); len(vals) > 0 {
			corrID = vals[0]
		}
	}
	msg := fmt.Sprintf(format, v...)
	log.Printf("[%s] %s", corrID, msg)
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
		logWithCorrelation(ctx, "[gRPC] Served %s from CACHE", cacheKey)

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
	logWithCorrelation(ctx, "[gRPC] Served %s from TCMB", cacheKey)

	// return success response
	return &pb.RateResponse{
		Currency: req.Currency,
		Date: req.Date,
		Rate: rate,
		Source: "TCMB",
	}, nil
}

// corresponding method in rpc for Go
func (s *GrpcServer) GetCryptoRate(ctx context.Context, req *pb.CryptoRequest) (*pb.CryptoResponse, error) {
	symbol := req.Symbol // e.g. "BTCUSDT"

	// cache control
	if cachedPrice, exists := s.CryptoCache.Get(symbol); exists {
		logWithCorrelation(ctx, "[gRPC] Served %s from CACHE", symbol)
		return &pb.CryptoResponse{
			Symbol: symbol,
			Price:  cachedPrice,
			Source: "Cache",
		}, nil
	}

	// if not in cache or TTL passed, fetch from Binance
	price, err := rates.FetchCryptoPrice(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto price: %v", err)
	}

	// write to cache and set 5 mins TTL
	s.CryptoCache.Set(symbol, price, 5)
	logWithCorrelation(ctx, "[gRPC] Served %s from BINANCE API", symbol)

	// return success response
	return &pb.CryptoResponse{
		Symbol: symbol,
		Price:  price,
		Source: "Binance API",
	}, nil
}