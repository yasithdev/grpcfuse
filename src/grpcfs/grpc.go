// place for grpc calls

package grpcfs

import (
	"context"
	pb "grpcfs/pb"
	"io/fs"
	"log"
)

// getting filesystem stats

func getStatFs(fsClient pb.FuseServiceClient, ctx context.Context, root string) (*pb.StatFs, error) {
	req := &pb.StatFsReq{
		Name:    root,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.StatFs(ctx, req)
	if err != nil {
		return nil, err
	}
	raw := res.Result
	if raw == nil {
		return nil, ctx.Err()
	}
	return raw, err
}

// getting file stats

func getStat(fsClient pb.FuseServiceClient, ctx context.Context, path string) (fs.FileInfo, error) {
	log.Print("grpc.getStat - path=", path)
	req := &pb.FileInfoReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	log.Print("grpc.getStat - calling fsClient.FileInfo for ", path)
	res, err := fsClient.FileInfo(ctx, req)
	if err != nil {
		log.Print("grpc.getStat - fsClient.FileInfo raised error. ", err)
		return nil, err
	}
	raw := res.Result
	if raw == nil {
		return nil, ctx.Err()
	}
	result := &FileInfoBridge{info: *raw}
	return result, err
	// direct command is - return os.Stat(path)
}

// getting directory entries

func readDir(fsClient pb.FuseServiceClient, ctx context.Context, path string) ([]fs.DirEntry, error) {
	req := &pb.ReadDirReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.ReadDir(ctx, req)
	raw := res.Result
	var entries []fs.DirEntry
	for _, entry := range raw {
		entries = append(entries, &DirEntryBridge{info: *entry})
	}
	return entries, err
	// direct command is - return os.ReadDir(path)
}

func readFile(fsClient pb.FuseServiceClient, ctx context.Context, path string) ([]byte, error) {
	req := &pb.ReadFileReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.ReadFile(ctx, req)
	return res.Result.Dst, err
	// direct command is - return os.ReadDir(path)
}
