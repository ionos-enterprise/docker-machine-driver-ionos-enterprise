
default: build

clean:
	$(RM) ./bin/docker-machine-driver-profitbricks
	$(RM) $(GOPATH)/bin/docker-machine-driver-profitbricks

build: clean
	GOGC=off go build -i -o ./bin/docker-machine-driver-profitbricks ./bin

install: build
	cp ./bin/docker-machine-driver-profitbricks $(GOPATH)/bin/

.PHONY: build install