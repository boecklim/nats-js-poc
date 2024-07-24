
.PHONY: run-rmi
run-rmi:
	docker-compose down && \
	(docker rmi nats-js-poc-pub || true) && \
	(docker rmi nats-js-poc-sub || true) && \
	docker-compose up --remove-orphans --scale pub=3 --scale sub=5

.PHONY: run
run:
	docker-compose down && \
	docker-compose up --remove-orphans --scale pub=3 --scale sub=5


.PHONY: run-nats
run-nats:
	docker-compose down && \
	docker-compose up --remove-orphans nats-a nats-b nats-c
