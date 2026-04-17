# Assignment 2 — gRPC Migration & Contract-First Development

## Overview
This project consists of two independent microservices: **Order Service** and **Payment Service**.

The main goal is to migrate internal communication from REST to **gRPC** and implement a **Contract-First approach** using Protocol Buffers.

---

## Repositories
- Proto Repository: https://github.com/zhettick/order-payment-protos.git
- Generated Code Repository: https://github.com/zhettick/order-payment-gen.git

---

## Architecture Overview

### Architecture Diagram
![Architecture Diagram](screenshots/diagram.png)

### Service Decomposition

#### Order Service
- Handles order lifecycle:
    - create order
    - get order
    - cancel order
    - get recent orders
- Exposes REST API (Gin)
- Acts as gRPC client (calls Payment Service)
- Acts as gRPC server (streams order updates)

#### Payment Service
- Processes payments
- Applies business rules
- Exposes gRPC server
- Includes logging interceptor

---

## Bounded Contexts

### Order Context
- order creation
- status management
- streaming updates

### Payment Context
- payment processing
- validation logic

Each service:
- has its own database
- contains its own business logic
- does not share internal code

---

## Clean Architecture

Each service follows layered architecture:

1. **Domain**
    - entities
    - repository interfaces

2. **Use Case**
    - business logic
    - validation rules

3. **Repository**
    - PostgreSQL implementation

4. **Transport**
    - HTTP (Order Service only)
    - gRPC (internal communication)
    - gRPC streaming

5. **Migrations**
    - SQL schema

---

## Inter-Service Communication

1. Client sends `POST /orders`
2. Order Service creates order (`Pending`)
3. Order Service calls Payment Service via gRPC
4. Payment Service processes payment
5. Order status becomes:
    - `Paid` (authorized)
    - `Failed` (declined)

---

## Contract-First Development

- `.proto` files define:
    - services
    - messages
    - enums
- Generated code stored in separate repository
- Both services import generated code

---

## Streaming

### RPC
`SubscribeToOrderUpdates`

### Behavior
- client subscribes by order ID
- Order Service checks database state
- when order status changes → update is sent via stream

---

## Error Handling

- gRPC timeout: `2s`
- Uses:
    - `status`
    - `codes`

Examples:
- `codes.InvalidArgument`
- `codes.NotFound`
- `codes.Internal`

### Failure Scenario
- Payment unavailable → HTTP `503`
- Order marked as `Failed`

---

## Additional Features

### gRPC Interceptor
Payment Service logs:
- method name
- execution time
- errors

---

## How to Run

### Docker
```
docker compose up --build
```

---

### Manual Run

#### Payment Service
```
cd payment
go run cmd/payment/main.go
```

#### Order Service
```
cd order
go run cmd/order/main.go
```

---

## Testing

### Create Order
```
POST http://localhost:8080/orders
```

```json
{
  "customer_id": "user_123",
  "item_name": "Laptop",
  "amount": 50000
}
```
![createOrderS.png](screenshots/createOrderS.png)
---

### Failed Payment
```json
{
  "customer_id": "user_123",
  "item_name": "Car",
  "amount": 150000
}
```
![createOrderF.png](screenshots/createOrderF.png)
---

### Get Order
```
GET /orders/{id}
```
![getOrder.png](screenshots/getOrder.png)
---

### Cancel Order
```
PATCH /orders/{id}/cancel
```
![cancelOrderF.png](screenshots/cancelOrderF.png)
---

### Streaming Test

```
evans --host localhost --port 50052 -r repl
package order.v1.service
service OrderService
call SubscribeToOrderUpdates
```
```sql
UPDATE orders SET status = 'Cancelled' WHERE id = '<order_id>';
```
![subscribe.png](screenshots/subscribe.png)
---
