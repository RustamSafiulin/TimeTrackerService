
executable := activity_service

all: $(executable)

clean:
	@rm -rf bin/* || true
	@rm -r bin || true

$(executable):
	@echo Get all dependecies for build
	go get -d .

	@echo Building $(executable)
	go build -o bin/$@ main.go

build:
	@echo Building $(executable)
	go build -o bin/$@ main.go

.PHONY: $(executable) clean build
