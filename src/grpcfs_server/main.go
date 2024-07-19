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

func handleErr(err error, message string) error {
	if err != nil {
		logger.Printf("%s: %v\n", message, err)
		return err
	}
	return nil
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
	logger.Print("received valid StatFS request. ", path, rpcCtx)
	return nil, nil
}

func (s *server) FileInfo(ctx context.Context, req *pb.FileInfoReq) (*pb.FileInfoRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid FileInfo request. ", path, rpcCtx)
	fileInfo, err := os.Stat(path)
	if handleErr(err, "os.Stat failed") != nil {
		return nil, err
	}
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
	logger.Print("responded valid FileInfo. ", res.Result)
	return res, nil
}

// TODO implement any locks here
func (s *server) OpenDir(ctx context.Context, req *pb.OpenDirReq) (*pb.OpenDirRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid OpenDir request. ", path, rpcCtx)
	res := &pb.OpenDirRes{
		Result: &pb.OpenedDir{},
	}
	return res, nil
}

// TODO implement any locks here
func (s *server) OpenFile(ctx context.Context, req *pb.OpenFileReq) (*pb.OpenFileRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid OpenFile request. ", path, rpcCtx)
	res := &pb.OpenFileRes{
		Result: &pb.OpenedFile{},
	}
	return res, nil
}

func (s *server) ReadDir(ctx context.Context, req *pb.ReadDirReq) (*pb.ReadDirRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid ReadDir request. ", path, rpcCtx)
	entries, err := os.ReadDir(path)
	if handleErr(err, "os.ReadDir failed") != nil {
		return nil, err
	}
	resEntries := []*pb.DirEntry{}
	for _, entry := range entries {
		info, err := entry.Info()
		if handleErr(err, "entry.Info() failed") != nil {
			return nil, err
		}
		obj := pb.DirEntry{
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			FileMode: uint32(entry.Type()),
			Info: &pb.FileInfo{
				Name:    info.Name(),
				Size:    info.Size(),
				Mode:    uint32(info.Mode()),
				ModTime: timestamppb.New(info.ModTime()),
				IsDir:   info.IsDir(),
				Ino:     info.Sys().(*syscall.Stat_t).Ino,
			},
		}
		resEntries = append(resEntries, &obj)
	}
	res := &pb.ReadDirRes{
		Result: resEntries,
	}

	return res, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileReq) (*pb.ReadFileRes, error) {
	path := req.Name
	rpcCtx := req.Context
	logger.Print("received valid ReadFile request. ", path, rpcCtx)
	file, err := os.ReadFile(path)
	if handleErr(err, "os.Stat failed") != nil {
		return nil, err
	}
	// Only Dst is used
	res := &pb.ReadFileRes{
		Result: &pb.FileEntry{
			Dst: file,
		},
	}
	return res, nil
}

func main() {

	listener, err := net.Listen("tcp", "127.0.0.1:50000")
	if handleErr(err, "Could not start GRPC server") != nil {
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterFuseServiceServer(s, &server{})

	go s.Serve(listener)
	logState("running until interrupt")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	logState("interrupt received, terminating.")
}
