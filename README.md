# 💱 Exchange Rate Microservice (Go)

A blazing-fast, lightweight, and thread-safe microservice built with Go. This service acts as an independent layer to fetch, cache, and serve historical fiat exchange rates from the Central Bank of the Republic of Turkey (TCMB) and real-time cryptocurrency prices via the Binance API. It features a dual-protocol architecture, communicating via both high-performance gRPC and standard REST, without bottlenecking the main application architecture.

## 🚀 Key Features

* **Dual-Protocol Architecture (gRPC & REST):** Concurrently runs a high-performance gRPC server (port 50051) for lightning-fast internal microservice communication and a standard REST API (port 8080), sharing the same underlying business logic and memory.
* **Real-time Cryptocurrency Tracking (Binance API):** Fetches live crypto-to-USD prices (e.g., BTCUSDT, ETHUSDT) on demand.
* **TTL-Based Crypto Caching:** Implements a specialized `sync.RWMutex` cache with a 5-minute Time-To-Live (TTL) expiration for highly volatile crypto assets, ensuring ultra-fast responses without hitting external API rate limits.
* **Proactive Fiat Background Worker:** A dedicated Goroutine wakes up automatically every workday at exactly 15:30 (TRT) to proactively fetch and cache highly used currencies (EUR, USD). Keeping CPU footprint at 0% while idle.
* **Smart Time-Travel Logic:** Implements a temporal fallback mechanism for fiat currencies. If a rate is requested for a weekend or an invalid future date, the algorithm automatically fetches the last valid workday's rate.
* **Minimalist & Cloud-Native:** Packaged using a multi-stage Docker build with an `alpine` base. The final compiled binary results in an ultra-lightweight Docker image (~15MB), currently deployed in a self-hosted cloud environment.

## ⚡ Performance Metrics (Live Prod Data)

By leveraging Go's raw performance, binary serialization (Protobuf), and a smart caching layer, the response times are drastically reduced:
* **Cold Boot (Network Fetch via TCMB/Binance):** ~273 ms
* **Cache Hit (RAM Retrieval via REST/gRPC):** ~145 µs *(Over 1800x faster!)*
* **gRPC Overhead:** Near-zero, offering heavily optimized binary data transfer compared to standard JSON parsing.

## 🛠️ Tech Stack

* **Language:** Go 1.26
* **Communication:** gRPC & Protocol Buffers (Protobuf), `go-chi/chi/v5` for REST
* **Concurrency:** Native Goroutines & Channels, `sync.RWMutex`
* **External APIs:** TCMB (Fiat), Binance API (Crypto)
* **Deployment:** Docker, Coolify (Self-hosted)

## 📡 Communication Interfaces

### 1. gRPC Service (Port: 50051)
The primary, high-performance interface for internal microservice communication.

**Service Definition:**
```protobuf
service ExchangeRateService {
  // Fiat Currency
  rpc GetExchangeRate (RateRequest) returns (RateResponse);
  
  // Cryptocurrency
  rpc GetCryptoRate (CryptoRequest) returns (CryptoResponse);
}
```

### 2. REST API (Port: 8080)
Available for standard HTTP clients, frontend testing, and fallback mechanisms.

**Endpoint:** `GET /api/rates`

**Query Parameters:**
* `currency` (string, required): Currency code (e.g., `USD`, `EUR`)
* `date` (string, required): Target date in `YYYY-MM-DD` format.

**Example Response:**
```json
{
  "currency": "USD",
  "date": "2026-07-04",
  "rate": 46.6337,
  "source": "TCMB"
}
```

## 🏗️ Local Development

1. Clone the repository.
2. Install dependencies: `go mod download`
3. *(Optional)* Regenerate Protobuf files if `rate.proto` is modified:
   `protoc --go_out=. --go-grpc_out=. proto/rate.proto`
4. Run the server: `go run cmd/api/main.go`
5. The REST API will be available at `http://localhost:8080` and the gRPC server at `:50051`.