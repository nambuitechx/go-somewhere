.PHONY: build push run down login

BACKEND_IMAGE = go-somewhere/backend
FORWARD_PROXY_IMAGE = go-somewhere/forward-proxy
REVERSE_PROXY_IMAGE = go-somewhere/reverse-proxy

TAG = 1.0.0
REGION = ap-southeast-1

ACCOUNT_ID = 832557411742
ECR_REGISTRY = $(ACCOUNT_ID).dkr.ecr.$(REGION).amazonaws.com

build-backend:
	docker build -t $(BACKEND_IMAGE):$(TAG) ./backend

# build-backend:
# 	docker build --platform linux/amd64 -t $(BACKEND_IMAGE):$(TAG) ./backend

build-forward-proxy:
	docker build -t $(FORWARD_PROXY_IMAGE):$(TAG) ./forward-proxy

# build-forward-proxy:
# 	docker build --platform linux/amd64 -t $(FORWARD_PROXY_IMAGE):$(TAG) ./forward-proxy

push-backend: login
	docker tag $(BACKEND_IMAGE):$(TAG) $(ECR_REGISTRY)/$(BACKEND_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(BACKEND_IMAGE):$(TAG)

push-forward-proxy: login
	docker tag $(FORWARD_PROXY_IMAGE):$(TAG) $(ECR_REGISTRY)/$(FORWARD_PROXY_IMAGE):$(TAG)
	docker push $(ECR_REGISTRY)/$(FORWARD_PROXY_IMAGE):$(TAG)
	
build: build-backend

push: push-backend

run:
	docker compose up -d

down:
	docker compose down

login:
	aws ecr get-login-password --region $(REGION) | docker login --username AWS --password-stdin $(ECR_REGISTRY)
