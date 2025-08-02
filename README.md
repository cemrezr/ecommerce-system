# 🛒 E-Commerce Order Processing System

A microservices-based e-commerce system built using event-driven architecture and RabbitMQ for inter-service communication. This system supports full order lifecycle management including stock updates and user notifications, with features like retry, dead letter queue, replay, and structured logging.

---

## 📐 Architecture Overview

```mermaid
graph TD
    subgraph RabbitMQ
        A[order.events (topic)] -->|order.created| B
        A -->|order.cancelled| B
        A -->|order.created| C
        A -->|order.cancelled| C
        A -->|order.created| D
        A -->|order.cancelled| D
    end

    subgraph Services
        B[Inventory Service]
        C[Notification Service]
        D[Order Service]
    end

    D -->|HTTP| User
