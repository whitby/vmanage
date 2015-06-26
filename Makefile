VERSION='v0.0.1'
all: clean build
clean:
	rm -f ${PWD}/out/vmanage-linux-amd64-${VERSION}
	rm -f ${PWD}/out/vmanage-darwin-amd64-${VERSION}
	rm -f ${PWD}/out/vmanage-linux-amd64-${VERSION}.tar.gz
	rm -f ${PWD}/out/vmanage-darwin-amd64-${VERSION}.tar.gz

build:
	docker build -t vmanage:build .	
	docker run -it --rm --name vmanage-build -e CGO_ENABLED=0 -e GOOS=linux -v ${PWD}/out:/var/shared vmanage:build go build -a -tags netgo -ldflags '-w' -o /var/shared/vmanage-linux-amd64-${VERSION}
	docker run -it --rm --name vmanage-build -e CGO_ENABLED=0 -e GOOS=darwin -v ${PWD}/out:/var/shared vmanage:build go build -a -tags netgo -ldflags '-w' -o /var/shared/vmanage-darwin-amd64-${VERSION}

zip:
	cd ${PWD}/out && tar -cvf vmanage-linux-amd64-${VERSION}.tar.gz vmanage-linux-amd64-${VERSION} && cd ..
	cd ${PWD}/out && tar -cvf vmanage-darwin-amd64-${VERSION}.tar.gz vmanage-darwin-amd64-${VERSION} && cd ..
