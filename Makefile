BIN = udpforward
ARCH = arm64
#ARCH = amd64
IMAGE = $(BIN)
VERSION = 0.2
DOCKERFILE = dockerfile
CONFIG = config.toml

build: $(IMAGE).tar
$(BIN): main.go
	CGO_ENABLE=0 GOOS=linux GOARCH=$(ARCH) go build -o $(BIN)  --tags=netgo,osusergo

$(IMAGE).tar: $(DOCKERFILE) $(BIN) $(CONFIG) .dockerignore
	sudo docker buildx build  -t $(IMAGE):$(VERSION) -f $(DOCKERFILE) --platform linux/$(ARCH) ./
	sudo docker save $(IMAGE):$(VERSION)>$(IMAGE).tar
	sudo docker rmi $(IMAGE):$(VERSION)

$(DOCKERFILE): generator/gen 
	generator/gen -b $(BIN) -d $(DOCKERFILE) -c $(CONFIG) -t generator/dockerfile.temp

generator/gen: generator/main.go generator/dockerfile.temp
	go build -o generator/gen generator/main.go

.dockerignore: Makefile
	echo > .dockerignore
	echo * >> .dockerignore
	echo !$(BIN) >> .dockerignore
	echo !$(CONFIG) >> .dockerignore

clean:
	-rm -f .dockerignore
	-rm -f generator/gen
	-rm -f $(DOCKERFILE)
	-rm -f $(IMAGE).tar
	-rm -f $(BIN)
