### Init
.PHONY: build push run down login

BACKEND_IMAGE = go-somewhere/backend
FORWARD_PROXY_IMAGE = go-somewhere/forward-proxy
REVERSE_PROXY_IMAGE = go-somewhere/reverse-proxy

REDIS_IMAGE = go-somewhere/redis
KAFKA_IMAGE = go-somewhere/kafka
SCYLLA_IMAGE = go-somewhere/scylla
PROMETHEUS_IMAGE=go-somewhere/prometheus

TAG = 1.0.0
REGION = ap-southeast-1

ACCOUNT_ID = 832557411742
ECR_REGISTRY = $(ACCOUNT_ID).dkr.ecr.$(REGION).amazonaws.com

### Build
# build-backend:
# 	docker build --platform linux/amd64 -t $(BACKEND_IMAGE):$(TAG) ./backend

build-backend:
	docker build -t $(BACKEND_IMAGE):$(TAG) ./backend

build-forward-proxy:
	docker build -t $(FORWARD_PROXY_IMAGE):$(TAG) ./forward-proxy

build-reverse-proxy:
	docker build -t $(REVERSE_PROXY_IMAGE):$(TAG) ./reverse-proxy

build-redis:
	docker build -t $(REDIS_IMAGE):$(TAG) ./redis

build-kafka:
	docker build -t $(KAFKA_IMAGE):$(TAG) ./kafka

build-scylla:
	docker build -t $(SCYLLA_IMAGE):$(TAG) ./scylla

build-prometheus:
	docker build -t $(PROMETHEUS_IMAGE):$(TAG) ./prometheus

build: build-backend build-forward-proxy build-reverse-proxy

### Push
push-backend: login
	docker tag $(BACKEND_IMAGE):$(TAG) $(ECR_REGISTRY)/$(BACKEND_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(BACKEND_IMAGE):$(TAG)

push-forward-proxy: login
	docker tag $(FORWARD_PROXY_IMAGE):$(TAG) $(ECR_REGISTRY)/$(FORWARD_PROXY_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(FORWARD_PROXY_IMAGE):$(TAG)

push-reverse-proxy: login
	docker tag $(REVERSE_PROXY_IMAGE):$(TAG) $(ECR_REGISTRY)/$(REVERSE_PROXY_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(REVERSE_PROXY_IMAGE):$(TAG)

push-redis: login
	docker tag $(REDIS_IMAGE):$(TAG) $(ECR_REGISTRY)/$(REDIS_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(REDIS_IMAGE):$(TAG)

push-kafka: login
	docker tag $(KAFKA_IMAGE):$(TAG) $(ECR_REGISTRY)/$(KAFKA_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(KAFKA_IMAGE):$(TAG)

push-scylla: login
	docker tag $(SCYLLA_IMAGE):$(TAG) $(ECR_REGISTRY)/$(SCYLLA_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(SCYLLA_IMAGE):$(TAG)

push-prometheus: login
	docker tag $(PROMETHEUS_IMAGE):$(TAG) $(ECR_REGISTRY)/$(PROMETHEUS_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(PROMETHEUS_IMAGE):$(TAG)

push: push-backend push-forward-proxy push-reverse-proxy

### Others
run:
	docker compose up -d

down:
	docker compose down

login:
	aws ecr get-login-password --region $(REGION) | docker login --username AWS --password-stdin $(ECR_REGISTRY)
