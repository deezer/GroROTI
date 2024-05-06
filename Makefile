BINARY=groroti

clean:
	rm -f data/qr/*.png

dev: clean
	SERVER_ADDR=127.0.0.1 SERVER_PORT=3000 go run main.go

snapshot:
	goreleaser --snapshot --clean

build:
	echo "Build the application as ./${BINARY}"
	go build -o ${BINARY} -ldflags "-extldflags '-static' -X main.Version=$$VERSION" main.go

dockerbuild: build
	docker build -t deezer/groroti:latest --build-arg VERSION=$$VERSION .
	docker build -t deezer/groroti:$$VERSION --build-arg VERSION=$$VERSION .

dockerrun:
	docker run -p 3000:3000 deezer/groroti:latest
