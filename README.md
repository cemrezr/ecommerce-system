# 🛒 E-Commerce Order Processing System

A microservices-based e-commerce system built using event-driven architecture and RabbitMQ for inter-service communication. This system supports full order lifecycle management including stock updates and user notifications, with features like retry, dead letter queue, replay, and structured logging.

---

Services
1. 🧾 Order Service
Exposes HTTP API to place or cancel orders.

Publishes order.created or order.cancelled events.

Logs events to PostgreSQL (event_logs).

Includes:

Retry logic

Circuit breaker

Dead Letter Queue (DLQ)

Replayer CLI tool (order-replayer) for failed events

2. 📦 Inventory Service
Listens to order events to update stock.

Updates inventory and logs in stock_logs.

Provides HTTP API to create/update products.

3. ✉️ Notification Service
Listens to order events and sends email notifications (simulated).

Handles order.created and order.cancelled events.

🔁 Event Flows
✅ Order Created
Client calls POST /orders.

Order Service:

Creates order

Publishes order.created

Logs event

Inventory Service:

Decreases product stock

Logs stock change

Notification Service:

Sends confirmation email

❌ Order Cancelled
Client calls POST /orders/{id}/cancel.

Order Service:

Cancels order

Publishes order.cancelled

Inventory Service:

Increases product stock

Notification Service:

Sends cancellation email

⚙️ Tech Stack
Layer	Technology
Language	Go (Golang)
Messaging	RabbitMQ (topic + DLQ)
Database	PostgreSQL
Logging	zerolog
Circuit	gobreaker
Container	Docker, Docker Compose
HTTP	Gorilla Mux

🚨 Error Handling Features
✅ Retry (up to 3 attempts per event)

✅ DLQ (order.failed queue + order.dlx exchange)

✅ Circuit breaker for RabbitMQ publishing

✅ Event replay CLI (order-replayer)

✅ Structured logging with correlation IDs (optional)

❌ (Optional) JSON Schema validation for event payloads

❌ (Optional) Idempotency key tracking

