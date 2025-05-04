# 🔐 Rate Limiter (Token Bucket Algorithm)
## 📌 Project Overview
This project implements a rate-limiting HTTP service using the Token Bucket algorithm to protect internal services from overuse and ensure fair distribution of resources between clients.

### The system allows:

- Per-client rate limits
- Individual refill rates and capacity
- Unlimited access for privileged clients
- Configuration via YAML file and dynamic updates via REST API
- Storage of client settings in PostgreSQL
- Concurrent-safe operations with minimal locks

# Getting started
1. Fill the config.yaml file. For example:
```yaml
default_limit:
  capacity: 100
  refill_rate_seconds: 1
client_rate_limits:
  - key: "user1"
    capacity: 50
    refill_rate_seconds: 5
    unlimited: false
  - key: "admin1"
    unlimited: true
address: :8080
log_level: DEBUG
db_host: db
db_user: postgres
db_password: postgres
db_name: ratelimiter
db_port: 5432
```
2. From root folder run
```sh
docker compose up --build
```


# Bot commands:

| Method   | Endpoint                | Description                          | Example |
|----------|-------------------------|--------------------------------------|---------|
| POST     | `/clients`              | Add a new client                     | `curl -X POST http://localhost:8080/clients -H "Content-Type: application/json" -d '{"client_id": "user1", "capacity": 2, "refill_rate_seconds": 1}'` |
| GET      | `/clients`              | List all clients                     | `curl http://localhost:8080/clients` |
| GET      | `/clients/{client_id}`        | Get a client by key                  | `curl http://localhost:8080/clients/{client_id}` |
| PUT      | `/clients/{client_id}`        | Update client's info                 | `curl -X PUT http://localhost:8080/clients/{client_id} -H "Content-Type: application/json" -d '{"capacity": 5, "refill_rate_seconds": 2}'` |
| DELETE   | `/clients/{client_id}`        | Delete a client                      | `curl -X DELETE http://localhost:8080/clients/{client_id}` |
| POST     | `/api`                  | Protected endpoint with rate limiting | `curl -H "X-API-Key: {client_id}" http://localhost:8080/api` |

- client_id - client id, specified as client_id while creating new user
- capacity - the maximum number of requests a client can make before hitting the limit.
- refill_rate_seconds - how often one token is added back to the bucket in seconds (e.g., `refill_rate_seconds = 1` means 1 new available token per second).

## Full testing pipeline:
1. After running the programm with docker compose create new user:
```sh
curl -X POST http://localhost:8080/clients   -H "Content-Type: application/json"   -d '{
    "client_id": "user1",
    "capacity": 11,
    "refill_rate_seconds": 1
  }'
```
2. Now you can make several request by this user:
```sh
for i in {1..15}; do      curl -H "X-API-Key: user1" http://localhost:8080/api; done
```
And see that only first 11 request end with "Request allowed", and every `refill_rate_seconds` seconds user1 will get one more allowed request
