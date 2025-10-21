### Init
.PHONY: build push run down login

BACKEND_IMAGE = go-somewhere/backend
FORWARD_PROXY_IMAGE = go-somewhere/forward-proxy
REVERSE_PROXY_IMAGE = go-somewhere/reverse-proxy
TEST_REVERSE_PROXY_IMAGE = go-somewhere/test-reverse-proxy

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

# build-test-reverse-proxy:
# 	docker build -t $(TEST_REVERSE_PROXY_IMAGE):$(TAG) ./test-reverse-proxy

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

push: push-backend push-forward-proxy push-reverse-proxy

### Others
run:
	docker compose up -d

down:
	docker compose down

login:
	aws ecr get-login-password --region $(REGION) | docker login --username AWS --password-stdin $(ECR_REGISTRY)
