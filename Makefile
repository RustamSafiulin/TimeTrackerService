
all: clean build

clean:
	make -C activity_service/ clean
	make -C mail_service/ clean

build:
	make -C activity_service/ build
	make -C mail_service/ build

