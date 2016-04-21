
default: build

clean:
	$(RM) ./bin/docker-machine-driver-profitbricks
	$(RM) $(GOPATH)/bin/docker-machine-driver-profitbricks

build: clean
	GOOS=darwin GOARCH=amd64 GOGC=off  CGOENABLED=0 go build -i -o ./bin/docker-machine-driver-profitbricks ./bin
	GOOS=windows GOARCH=amd64 GOGC=off CGOENABLED=0 go build -i -o ./bin/docker-machine-driver-profitbricks.exe ./bin
	GOOS=linux GOARCH=amd64 GOGC=off CGOENABLED=0 go build -i -o ./bin/docker-machine-driver-profitbricks ./bin

install: build
	cp ./bin/docker-machine-driver-profitbricks $(GOPATH)/bin/

.PHONY: build install