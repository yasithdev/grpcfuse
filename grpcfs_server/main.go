package main

import (
	"context"
	"flag"
	"grpcfs/pb"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	"google.golang.org/grpc"
)

var logger = log.Default()

func handleErrIfAny(err error, message string) {
	if err != nil {
		logger.Fatalf("%s: %v\n", message, err)
	}
}

func logState(message string, v ...any) {
	logger.Print(message, v)
}

type server struct {
	pb.FuseServiceServer
}

func (s *server) StatFs(ctx context.Context, req *pb.StatFsReq) (*pb.StatFsRes, error) {
	return nil, nil
}

func (s *server) FileInfo(ctx context.Context, req *pb.FileInfoReq) (*pb.FileInfoRes, error) {
	return nil, nil
}

func (s *server) OpenDir(ctx context.Context, req *pb.OpenDirReq) (*pb.OpenDirRes, error) {
	return nil, nil
}

func (s *server) OpenFile(ctx context.Context, req *pb.OpenFileReq) (*pb.OpenFileRes, error) {
	return nil, nil
}

func (s *server) ReadDir(ctx context.Context, req *pb.ReadDirReq) (*pb.ReadDirRes, error) {
	return nil, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileReq) (*pb.ReadFileRes, error) {
	return nil, nil
}

func main() {

	var servePath string
	flag.StringVar(&servePath, "serve", "", "Path to serve")
	flag.Parse()

	if servePath == "" {
		logger.Fatal("Please specify which path to serve")
	}

	servePath, err := filepath.Abs(servePath)
	handleErrIfAny(err, "Invalid serve path")

	listener, err := net.Listen("tcp", ":8080")
	handleErrIfAny(err, "Could not start GRPC server")

	s := grpc.NewServer()
	pb.RegisterFuseServiceServer(s, &server{})

	go s.Serve(listener)
	logState("running until interrupt")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	logState("interrupt received, terminating.")
}
