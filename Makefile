
.PHONY: run
run:
	docker-compose down && \
	(docker rmi nats-js-poc-pub || true) && \
	(docker rmi nats-js-poc-sub || true) && \
	docker-compose up --scale pub=2 --scale sub=3


.PHONY: run-nats
run-nats:
	docker-compose down && \
	docker-compose up nats
