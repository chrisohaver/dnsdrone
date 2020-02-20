
DOCKER:=chrisohaver
all:
	docker buildx use dnsdrone-builder || docker buildx create --use --name dnsdrone-builder
	docker buildx build -t $(DOCKER)/dnsdrone --platform=linux/amd64,linux/arm,linux/arm64 . --push