package main

import (
	"context"
	"grpcfs/pb"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid StatFS request", path, rpcCtx)
	return nil, nil
}

func (s *server) FileInfo(ctx context.Context, req *pb.FileInfoReq) (*pb.FileInfoRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid FileInfo request", path, rpcCtx)
	fileInfo, err := os.Stat(path)
	handleErrIfAny(err, "cannot get FileInfo")
	stat := fileInfo.Sys().(*syscall.Stat_t)
	res := &pb.FileInfoRes{
		Result: &pb.FileInfo{
			Name:    fileInfo.Name(),
			Size:    fileInfo.Size(),
			Mode:    uint32(fileInfo.Mode()),
			ModTime: timestamppb.New(fileInfo.ModTime()),
			IsDir:   fileInfo.IsDir(),
			Ino:     stat.Ino,
		},
	}
	logger.Print("responded valid FileInfo", res.Result)
	return res, nil
}

func (s *server) OpenDir(ctx context.Context, req *pb.OpenDirReq) (*pb.OpenDirRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid OpenDir request", path, rpcCtx)
	// TODO check result
	res := &pb.OpenDirRes{
		Result: &pb.OpenedDir{},
	}
	return res, nil
}

func (s *server) OpenFile(ctx context.Context, req *pb.OpenFileReq) (*pb.OpenFileRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid OpenFile request", path, rpcCtx)
	// TODO check result
	res := &pb.OpenFileRes{
		Result: &pb.OpenedFile{},
	}
	return res, nil
}

func (s *server) ReadDir(ctx context.Context, req *pb.ReadDirReq) (*pb.ReadDirRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid ReadDir request", path, rpcCtx)
	// TODO check result
	res := &pb.ReadDirRes{
		Result: []*pb.DirEntry{{}},
	}
	return res, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileReq) (*pb.ReadFileRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid ReadFile request", path, rpcCtx)
	// TODO check result
	res := &pb.ReadFileRes{
		Result: &pb.FileEntry{},
	}
	return res, nil
}

func main() {

	listener, err := net.Listen("tcp", "127.0.0.1:50000")
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
