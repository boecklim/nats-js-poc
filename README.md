# NATS JetStream proof of concept

## How to run

```
make run
```

In order to build the app before running 
```
make run-rmi
```

Add a 4th nats server

```
docker-compose up --remove-orphans nats-d
```

Run benchmark test with JetStream, 50k messages of size 200 bytes, in memory storage, 5 publishers, 5 subscribers

```
nats bench foo --pub 5 --sub 5 --size 200 --js --pull --msgs=50000 --storage=memory
```

## Todos

- [X] Server configuration
  - [X] Clustering
  - [ ] Accounts
- [X] Connection configuration
  - [X] Reconnection
  - [X] Ping
- [X] App
  - [X] Subscriber
  - [X] Publisher
  - [X] Stream
  - [X] Consumer
