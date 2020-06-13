install:
	GO114MODULE=on go mod tidy

build: install
	GO114MODULE=on go build -o name-node.bin .

run: build
	./name-node.bin
