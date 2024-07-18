// place for grpc calls

package grpcfs

import (
	"context"
	pb "grpcfs/pb"
	"io/fs"
)

func getStatFs(fsClient pb.GrpcFuseClient, ctx context.Context, root string) (*pb.StatFsRes, error) {
	req := &pb.StatFsReq{
		Name:    root,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.StatFs(ctx, req)
	return res, err
}

func getStat(fsClient pb.GrpcFuseClient, ctx context.Context, path string) (fs.FileInfo, error) {
	req := &pb.StatReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.Stat(ctx, req)
	return res, err
	// direct command is - return os.Stat(path)
}

func readDir(fsClient pb.GrpcFuseClient, ctx context.Context, path string) ([]fs.DirEntry, error) {
	req := &pb.ReadDirReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.ReadDir(ctx, req)
	return res, err
	// direct command is - return os.ReadDir(path)
}

func readFile(fsClient pb.GrpcFuseClient, ctx context.Context, path string) ([]byte, error) {
	req := &pb.ReadFileReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.ReadFile(ctx, req)
	return res, err
	// direct command is - return os.ReadDir(path)
}
