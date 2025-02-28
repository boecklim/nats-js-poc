
.PHONY: run-rmi
run-rmi:
	docker-compose down && \
	(docker rmi nats-js-poc-pub || true) && \
	(docker rmi nats-js-poc-sub || true) && \
	docker-compose up --remove-orphans pub sub nats-a nats-b nats-c --scale pub=2 --scale sub=2

.PHONY: run
run:
	docker-compose down && \
	docker-compose up --remove-orphans pub sub nats-a nats-b nats-c --scale pub=2 --scale sub=2


.PHONY: run-nats
run-nats:
	docker-compose down && \
	docker-compose up --remove-orphans nats-a nats-b nats-c nats-d
