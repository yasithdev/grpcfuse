all: init client server

init:
	rm -rf ./bin && mkdir ./bin

client:
	cd src/grpcfs_client && go mod tidy && go build -o ../../bin .

server:
	cd src/grpcfs_server && go mod tidy && go build -o ../../bin .

run_server:
	bin/server

run_client:
	bin/client -mount tmp/ -serve data/