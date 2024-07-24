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

## Todos
- [X] Server configuration
    - [X] Clustering
- [X] Connection configuration
    - [X] Reconnection
    - [X] Ping
