# High-Level Design (HLD) - GoComet Ride-Hailing Application

## 1. System Overview

GoComet is a production-grade, multi-tenant ride-hailing platform designed to handle:
- **100,000 drivers** with real-time location tracking
- **10,000 ride requests per minute** (166 requests/second)
- **200,000 location updates per second** (2 updates/driver/second)
- **Sub-second driver matching** (p95 < 1s)

## 2. Architecture

### 2.1 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │ Rider UI     │  │  Driver UI   │  │  Admin Panel    │  │
│  │ (Web/Mobile) │  │  (Web/Mobile)│  │                 │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬────────┘  │
└─────────┼───────────────────┼──────────────────┼───────────┘
          │                   │                   │
          └───────────────────┴───────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   Load Balancer   │
                    │  (nginx/HAProxy)  │
                    └─────────┬─────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
    ┌─────▼─────┐      ┌─────▼─────┐      ┌─────▼─────┐
    │  API      │      │  API      │      │  API      │
    │  Server 1 │      │  Server 2 │      │  Server N │
    │ (Stateless)│     │ (Stateless)│     │ (Stateless)│
    └─────┬─────┘      └─────┬─────┘      └─────┬─────┘
          │                   │                   │
          └───────────────────┼───────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
    ┌─────▼─────┐      ┌─────▼─────┐      ┌─────▼─────┐
    │PostgreSQL │      │   Redis   │      │ New Relic │
    │  Primary  │      │  Cluster  │      │    APM    │
    │           │      │           │      │           │
    └───────────┘      └───────────┘      └───────────┘
          │
    ┌─────▼─────┐
    │PostgreSQL │
    │  Replica  │
    └───────────┘
```

### 2.2 Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
├─────────────────────────────────────────────────────────────┤
│  API Handlers (HTTP/WebSocket)                              │
│  ├── Ride Handler                                           │
│  ├── Driver Handler                                         │
│  ├── Trip Handler                                           │
│  ├── Payment Handler                                        │
│  └── WebSocket Handler                                      │
├─────────────────────────────────────────────────────────────┤
│  Middleware                                                  │
│  ├── Authentication & Authorization                         │
│  ├── Idempotency Check                                      │
│  ├── Rate Limiting                                          │
│  └── New Relic Instrumentation                             │
├─────────────────────────────────────────────────────────────┤
│  Business Logic (Services)                                   │
│  ├── Matching Service (Driver-Rider)                       │
│  ├── Pricing Service (Fare Calculation)                    │
│  ├── Location Service (Tracking)                           │
│  └── Notification Service (Real-time Updates)              │
├─────────────────────────────────────────────────────────────┤
│  Domain Layer                                                │
│  ├── Driver Entity & Repository Interface                  │
│  ├── Rider Entity & Repository Interface                   │
│  ├── Ride Entity & Repository Interface                    │
│  ├── Trip Entity & Repository Interface                    │
│  └── Payment Entity & Repository Interface                 │
├─────────────────────────────────────────────────────────────┤
│  Data Access Layer                                           │
│  ├── PostgreSQL Repositories                               │
│  └── Redis Cache Repositories                              │
└─────────────────────────────────────────────────────────────┘
```

## 3. Key Design Decisions

### 3.1 Clean Architecture
- **Separation of Concerns**: Clear boundaries between layers
- **Dependency Inversion**: Domain layer independent of infrastructure
- **Testability**: Each layer can be tested independently

### 3.2 Technology Choices

| Component | Technology | Justification |
|-----------|-----------|---------------|
| Language | Go | High performance, concurrency, low latency |
| Database | PostgreSQL | ACID compliance, complex queries, reliability |
| Cache | Redis | Geo-spatial commands, pub/sub, high throughput |
| Web Framework | Gin | Lightweight, fast routing, middleware support |
| WebSocket | Gorilla | Mature library, production-ready |
| Monitoring | New Relic | Comprehensive APM, distributed tracing |
| Logging | Zap | Structured logging, high performance |

### 3.3 Scalability Strategy

**Horizontal Scaling**
- Stateless API servers behind load balancer
- Connection pooling for database and Redis
- WebSocket scaling via Redis Pub/Sub

**Data Partitioning**
- Geographic sharding for multi-region support
- Read replicas for database scaling
- Redis clustering by key prefix

**Caching Strategy**
- Read-through cache for ride data
- Write-through cache for driver locations
- TTL-based invalidation

## 4. Data Flow Diagrams

### 4.1 Ride Request Flow (Synchronous Matching)

```
Rider -> POST /v1/rides -> API Server
           ↓
  [Validate Request]
           ↓
  [Query Redis GEORADIUS (nearby drivers within 5km)]
           ↓
  [Filter by availability & vehicle type]
           ↓
  [Select nearest available driver]
           ↓
  [Store ride in PostgreSQL (status: assigned, driver_id)]
           ↓
  [Send WebSocket notification to matched driver]
           ↓
  [Return ride details with matched driver to rider]
```

**Key Points:**
- Matching is **synchronous** - rider waits for driver assignment
- Response includes matched driver details immediately
- If no driver available, returns status "requested" for retry
- Sub-second matching via Redis GEORADIUS (O(log N) complexity)
```

### 4.2 Location Update Flow

```
Driver -> POST /v1/drivers/{id}/location -> API Server
             ↓
  [Rate Limit Check (2/sec)]
             ↓
  [GEOADD to Redis drivers:locations]
             ↓
  [SADD to Redis drivers:available (if online)]
             ↓
  [Async UPDATE PostgreSQL (debounced)]
             ↓
  [Broadcast location via WebSocket]
             ↓
  [Record metric in New Relic]
```

### 4.3 Trip Completion Flow

```
Driver -> POST /v1/trips/{id}/end -> API Server
            ↓
  [BEGIN Transaction]
            ↓
  [Calculate Fare]
    - Base fare by vehicle type
    - Distance fare (km × rate)
    - Time fare (min × rate)
    - Surge multiplier (from Redis)
    - Total = subtotal × surge
            ↓
  [UPDATE trip (ended_at, fare, status)]
            ↓
  [UPDATE ride (status: completed)]
            ↓
  [UPDATE driver (status: online)]
            ↓
  [CREATE payment record]
            ↓
  [COMMIT Transaction]
            ↓
  [Clear Redis cache]
            ↓
  [Send notifications]
            ↓
  [Record business metrics]
```

## 5. Performance Optimizations

### 5.1 Database Optimizations
- **Connection Pooling**: Max 100 connections, idle 10
- **Indexes**:
  - `(rider_id, status, created_at DESC)` - Active rides by rider
  - `(driver_id, status, created_at DESC)` - Active rides by driver
  - `(status, created_at DESC)` WHERE status IN ('requested', 'assigned')
  - `(status)` WHERE status = 'online' - Available drivers
- **Prepared Statements**: All queries use prepared statements
- **Query Optimization**: EXPLAIN ANALYZE for all queries

### 5.2 Caching Optimizations
- **Driver Locations**: GEORADIUS for O(log N) nearest neighbor search
- **Active Rides**: Hash cache with 5-minute TTL
- **Idempotency**: 24-hour TTL for duplicate prevention
- **Surge Pricing**: Region-based multipliers updated every 5 minutes

### 5.3 API Optimizations
- **Synchronous Matching**: Fast in-memory Redis lookups (<500ms p95)
- **Location Batching**: Group updates before writing to PostgreSQL
- **WebSocket**: Eliminate polling, push-based updates
- **Compression**: gzip for large responses

## 6. Reliability & Fault Tolerance

### 6.1 Data Consistency
- **ACID Transactions**: Multi-step operations wrapped in transactions
- **Optimistic Locking**: Version fields for concurrent updates
- **Pessimistic Locking**: SELECT FOR UPDATE for critical sections
- **Idempotency**: Prevent duplicate ride requests and payments

### 6.2 Error Handling
- **Graceful Degradation**: Fallback to PostgreSQL if Redis fails
- **Retry Logic**: Exponential backoff for transient failures
- **Circuit Breakers**: Stop cascading failures
- **Timeout Management**: All operations have timeouts

### 6.3 Monitoring & Alerting
- **APM**: Request traces, slow query detection
- **Custom Metrics**:
  - `custom/ride/matching_latency_ms`
  - `custom/driver/location_update_rate`
  - `custom/pricing/surge_multiplier`
  - `custom/db/connection_pool_usage`
- **Alerts**:
  - API latency p95 > 1s
  - Database connections > 80%
  - Redis memory > 80%
  - Error rate > 1%

## 7. Security

### 7.1 Authentication & Authorization
- JWT-based authentication (simplified for demo)
- Role-based access control (rider, driver, admin)
- API key rotation

### 7.2 Data Protection
- HTTPS/TLS encryption in transit
- Input validation and sanitization
- SQL injection prevention (parameterized queries)
- Rate limiting to prevent abuse

### 7.3 Compliance
- PII data encryption
- Audit logging for state changes
- GDPR-compliant data deletion

## 8. Scalability Targets

| Metric | Target | Strategy |
|--------|--------|----------|
| Drivers | 100,000 | Redis geo-spatial indexing |
| Ride Requests | 10,000/min | Horizontal API scaling |
| Location Updates | 200,000/sec | Redis write optimization |
| Matching Latency | <1s p95 | In-memory cache lookups |
| API Latency | <500ms p95 | Caching, Redis in-memory lookups |
| Availability | 99.9% | Multi-AZ deployment |

## 9. Future Enhancements

1. **Multi-Region Support**: Geographic data partitioning
2. **ML-Based Matching**: Predictive driver assignment
3. **Dynamic Pricing**: Real-time demand/supply analysis
4. **Route Optimization**: Google Maps/OSRM integration
5. **Fraud Detection**: ML-based anomaly detection
6. **Push Notifications**: FCM/APNs integration
7. **Analytics Platform**: Real-time business intelligence
8. **A/B Testing Framework**: Feature rollout control

## 10. Deployment Architecture

### 10.1 Infrastructure
- **Kubernetes**: Container orchestration
- **Docker**: Containerization
- **AWS/GCP**: Cloud provider
- **RDS/CloudSQL**: Managed PostgreSQL
- **ElastiCache**: Managed Redis

### 10.2 CI/CD Pipeline
- **GitHub Actions**: Automated testing and deployment
- **Docker Build**: Multi-stage builds
- **Helm Charts**: Kubernetes deployment
- **Blue-Green Deployment**: Zero-downtime releases

### 10.3 Disaster Recovery
- **Database Backups**: Daily snapshots, point-in-time recovery
- **Redis Persistence**: AOF for durability
- **Multi-AZ**: Automatic failover
- **RPO**: <1 hour
- **RTO**: <15 minutes

## 11. Cost Optimization

- **Auto-scaling**: Scale down during off-peak hours
- **Reserved Instances**: For predictable workloads
- **Spot Instances**: For batch processing
- **CDN**: CloudFront for static assets
- **Compression**: Reduce bandwidth costs

---

**Document Version**: 1.0
**Last Updated**: January 2026
**Author**: GoComet DAW Assessment
