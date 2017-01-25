IMAGE=containersol/geoserver

build:
	go get github.com/boltdb/bolt/...
	go build

docker:
	# docker run --rm -v "${PWD}":/usr/src/myapp -w /usr/src/myapp -e CGO_ENABLED=0 -e GOOS=linux golang:1.7.4 go build -a -installsuffix cgo -o geoserver .
	docker build -t ${IMAGE}:latest .	

push: build
	docker push ${IMAGE}

all: build
