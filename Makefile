.phony: build clean setup

run: build bash

build:
	go build

bash:
	sudo ./lizrice-container run /bin/bash

setup:
	docker run -d --rm --name ubuntufs ubuntu:20.04 sleep 1000
	docker export ubuntufs -o ubuntufs.tar
	docker stop ubuntufs
	mkdir ./ubuntufs
	tar xf ubuntufs.tar -C ./ubuntufs/
	touch ./ubuntufs/CONTAINER_FS

clean:
	rm -rf ubuntufs ubuntufs.tar lizrice-container
