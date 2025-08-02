# 🛒 E-Commerce Order Processing System

A microservices-based e-commerce system built using event-driven architecture and RabbitMQ for inter-service communication. This system supports full order lifecycle management including stock updates and user notifications, with features like retry, dead letter queue, replay, and structured logging.

---

## 📦 Services Overview

| Service             | Description                                   |
|---------------------|-----------------------------------------------|
| **Order Service**    | Handles order creation and cancellation       |
| **Inventory Service**| Manages product stock levels                 |
| **Notification Service** | Sends notifications (e.g. order confirmation) |

All services are containerized with **Docker** and use **RabbitMQ** for async messaging. **PostgreSQL** is used for persistent data storage and event logging.

---

## 🔁 Core Event Flows

### 1. Order Creation
- HTTP POST to `order-service` → validates stock via `inventory-service`
- Creates order in DB and publishes `order.created`
- `inventory-service` consumes event → decrements stock
- `notification-service` sends confirmation via `order.created`

### 2. Order Cancellation
- HTTP POST to `order-service` → checks existence
- Marks order as cancelled → publishes `order.cancelled`
- `inventory-service` consumes event → restores stock
- `notification-service` sends cancellation alert

---

## 📨 Event Topics (RabbitMQ)

| Event Type         | Description                    |
|--------------------|--------------------------------|
| `order.created`     | Order successfully created     |
| `order.cancelled`   | Order cancelled                |
| `order.failed`      | Event publishing failed → DLQ |

Exchange Type: `topic`  
Exchange Name: `order.events`  
DLQ Exchange: `order.dlx` → Queue: `order.failed`

---

## ✅ Features 

### 🔄 **Message Processing**
- **Retries:** All publishers retry 3 times on failure
- **DLQ:** Failed messages are routed to `order.failed` queue
- **Idempotency:** `inventory-service` prevents double processing via `stock_logs`
- **Ordering:** Handled via event timestamps (FIFO queues)

### 🧠 **Event Handling**
- **Validation:** Incoming HTTP payloads validated (type, constraints)
- **Versioning:** `event_version: v1` added to all events
- **Storage:** All events are logged in PostgreSQL `event_logs`
- **Replay:** Failed events retried via `order-replayer` tool

### 🔥 **Error Handling**
- **Circuit Breakers:** Applied to:
    - `order-service → RabbitMQ publisher`
    - `order-service → inventory-service HTTP client`
- **Connection Loss:** Handled via retry logic + exponential backoff
- **Invalid Formats:** All inputs validated with detailed error responses

---


## 🧪 Testing Strategy

### Manual Tests
You can use `curl`, Postman or HTTP clients to verify flows.


#### Create Product
```bash

curl --location --request POST 'http://localhost:8082/products' \
--header 'Content-Type: application/json' \
--data-raw '{
  "product_name": "iPhone",
  "stock": 50
}'
```

#### Create Order
```bash

curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "product_id": 1, "quantity": 1}'
```

#### Cancel Order
```bash

curl -X POST http://localhost:8081/orders/1/cancel
```

#### Replay Failed Events
```bash

cd order-service
make replay
```

### Error Scenarios Tested
- Service unavailability (e.g. kill inventory service)
- Broken DB connection
- Circuit breaker tripping
- Double event delivery (idempotency validation)

---

## 🐳 Infrastructure Setup

### Prerequisites
- [Docker + Docker Compose](https://docs.docker.com/get-docker/)
- Go 1.21+
- `make` utility

### Quick Start
```bash

# Start dependencies
docker-compose up -d

# Create DB schema (Postgres)
cd services/order-service && make migrate-up
cd services/inventory-service && make migrate-up

# Run services manually
cd services/order-service && make run
cd services/inventory-service && make run
cd services/notification-service && make run
```

---
