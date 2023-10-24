.PHONY: build-worker build-master

run-master: build-master
	./bin/master

run-worker: build-worker
	./bin/worker

build-all: build-master build-worker

build-master:
	go build -o bin/master ./cmd/master

build-worker:
	go build -o bin/worker ./cmd/worker
