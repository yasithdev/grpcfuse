all: init client server

init:
	rm -rf ./bin && mkdir ./bin

client:
	cd grpcfs_client && go mod tidy && go build -o ../bin .

server:
	cd grpcfs_server && go mod tidy && go build -o ../bin .