
executable := activity_service

all: $(executable)

clean:
	@rm -rf bin/* || true
	@rm -r bin || true

$(executable):
	@echo Get all dependencies of $(executable)
	go get -d .

	@echo Building $(executable)
	go build -o bin/$(executable) .

build:
	@echo Get all dependencies of $(executable)
	go get -d .

	@echo Building $(executable)
	go build -o bin/$(executable) .

.PHONY: $(executable) clean build
