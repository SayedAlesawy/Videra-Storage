install:
	GO114MODULE=on go mod tidy

build: install
	GO114MODULE=on go build -o data-node.bin .

run: build
	./data-node.bin
