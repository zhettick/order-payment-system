## Architecture Overview
The project consists of two independent microservices: **Order Service** and **Payment Service**. The core evolution in this version is the migration from **weak REST** contracts to **Strongly Typed gRPC contracts** and **Contract-First Development**.

### Architecture Diagram
![diagram.png](screenshots/diagram.png)

### Service Decomposition
- **Order Service:**
  - Handles order lifecycle (create, get, cancel)
  - Exposes REST API for external clients
  - Acts as a gRPC client when calling Payment Service
  - Acts as a gRPC server for streaming order updates
- **Payment Service:** 
  - Processes payments and validates limits
  - Exposes gRPC server
  - Implements business rules for authorization/decline
  - Includes gRPC interceptor (logging middleware)

### Bounded Contexts
The system follows Domain-Driven Design with two independent contexts:
1. Order Context 
   - order creation
   - status management
   - streaming updates
2. Payment Context
   - payment processing
   - transaction validation

Each service:
- owns its own database
- contains its own business logic
- does NOT share internal code

### Clean Architecture Layers
Each service follows a layered architecture:
1.  **Domain (Entities):** Pure Go structures without any external dependencies (no JSON/DB tags).
2.  **Use Case:** Contains business logic
     - validating order amount
     - enforcing order status transitions
     - applying payment rules
3.  **Repository:** Responsible for data persistence using PostgreSQL.
4.  **Transport (HTTP):** 
    - HTTP(Gin) → external API (Order Service only)
    - gRPC → Payment Service → gRPC Server; Order Service → gRPC Client + Streaming Server 
5.  **Migrations:** SQL scripts for database schema management.

### Inter-Service Communication
1.	Client sends POST /orders
2.	Order Service creates order with **"Pending"**
3.	Order Service calls Payment Service via gRPC
4.	Payment Service processes payment:
   - Authorized → Order becomes **"Paid"** 
   - Declined → Order becomes **"Failed"*

### Contract-First Development
The system follows Contract-First design:
- .proto files define:
  - services
  - messages
  - enums
- Generated Go code (pb.go) is used by both services

Structure:
- Repository A (Protos):
  - contains .proto definitions
- Repository B (Generated Code):
  - contains generated .pb.go
  - used via: ```go get github.com/zhettick/order-payment-gen```
  
### Resilience and Failure Handling
- gRPC client uses timeout (2 seconds)
- Failure scenario:
  - Payment Service unavailable → Order marked as **"Failed"**
  - HTTP returns **503 Service Unavailable**

To handle errors the system uses:
- google.golang.org/grpc/status
- codes.*

Examples:
- codes.InvalidArgument
- codes.NotFound
- codes.Internal

### Architecture Decisions
1. **Contract-First Approach**  
   The system uses .proto files to define all service contracts before implementation. This ensures strong typing and prevents breaking changes between services.
2. **gRPC instead of REST(Internal Communication)**  
   Internal service communication uses gRPC instead of REST for better performance and strict schema enforcement. It also enables efficient binary communication over HTTP/2.
3. **Database per Service**  
   Each microservice has its own database, ensuring complete data ownership. This improves isolation and allows services to scale independently.
4. **No Shared Domain Models**  
   Services do not share domain models to avoid tight coupling. All communication happens through generated gRPC contracts instead of shared code.
5. **Streaming via Channels**  
   Order updates are delivered in real-time using gRPC streaming combined with Go channels. This enables event-driven, immediate status updates.
6. **Manual Dependency Injection**  
   All dependencies are manually wired in main.go, which acts as the composition root. This keeps architecture explicit and easy to test.
7. **Layer Architecture**  
   Clear separation of responsibilities improves maintainability and testability.
---

## How to Run

### 1. Database Setup (Docker)
Ensure you have Docker installed and run:
```bash
docker-compose up -d
```

### 2. Run Services (if not using Docker for app)
Open two terminal tabs:

**Payment Service:**
```bash
cd payment
go run cmd/payment/main.go
```

**Order Service:**
```bash
cd order
go run cmd/order/main.go
```

---

## Testing the System

### Create Order (REST → triggers gRPC)
`POST http://localhost:8080/orders`
```json
{
    "customer_id": "user_123",
    "item_name": "Laptop",
    "amount": 50000
}
```
![createOrderS.png](screenshots/createOrderS.png)

### Create Order (Fail - Over Limit)
`POST http://localhost:8080/orders`
```json
{
    "customer_id": "user_123",
    "item_name": "Car",
    "amount": 150000
}
```
![createOrderF.png](screenshots/createOrderF.png)

### Get Order
`GET http://localhost:8080/orders/{id}`
![getOrder.png](screenshots/getOrder.png)

### Cancel Order
`PATCH http://localhost:8080/orders/{id}/cancel`
![cancelOrderF.png](screenshots/cancelOrderF.png)

### Streaming Test (gRPC)
Using Evans CLI: `evans -r repl`
Call: `SubscribeToOrderUpdates`
![createPayment.png](screenshots/createPayment.png)

``````
