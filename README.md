# NATS JetStream proof of concept

## How to run

```
docker-compose down && \ 
(docker rmi nats-js-poc-pub || true) && \
(docker rmi nats-js-poc-sub || true) && \
docker-compose up --abort-on-container-exit --scale pub=2 --scale sub=3
```

## Todo
- NATS server configuration
    - clustering
- NATS connection configuration
    - Reconnection
    - Ping
