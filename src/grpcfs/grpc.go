// place for grpc calls

package grpcfs

import (
	"context"
	pb "grpcfs/pb"
	"io/fs"
	"time"
)

// getting filesystem stats

func getStatFs(fsClient pb.FuseServiceClient, ctx context.Context, root string) (*pb.StatFs, error) {
	req := &pb.StatFsReq{
		Name:    root,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.StatFs(ctx, req)
	raw := res.Result
	if raw == nil {
		return nil, ctx.Err()
	}
	return raw, err
}

// getting file stats

type FileInfoBridge struct {
	info pb.FileInfo
}

func (b *FileInfoBridge) Name() string {
	return b.info.Name
}

func (b *FileInfoBridge) Size() int64 {
	return b.info.Size
}

func (b *FileInfoBridge) Mode() fs.FileMode {
	return fs.FileMode(b.info.Mode)
}

func (b *FileInfoBridge) ModTime() time.Time {
	return b.info.ModTime.AsTime()
}

func (b *FileInfoBridge) IsDir() bool {
	return b.info.IsDir
}

func (b *FileInfoBridge) Sys() any {
	return b.info.Sys
}

func getStat(fsClient pb.FuseServiceClient, ctx context.Context, path string) (fs.FileInfo, error) {
	req := &pb.FileInfoReq{
		Name:    path,
		Context: &pb.RPCContext{},
	}
	res, err := fsClient.FileInfo(ctx, req)
	raw := res.Result
	if raw == nil {
		return nil, ctx.Err()
	}
	result := &FileInfoBridge{info: *raw}
	return result, err
	// direct command is - return os.Stat(path)
}

// getting directory entries

type DirEntryBridge struct {
	info pb.DirEntry
}

func (b *DirEntryBridge) Name() string {
	return b.info.Name
}

func (b *DirEntryBridge) IsDir() bool {
	return b.info.IsDir
}

func (b *DirEntryBridge) Type() fs.FileMode {
	return fs.FileMode(b.info.FileMode)
}

func (b *DirEntryBridge) Info() (fs.FileInfo, error) {
	info := &FileInfoBridge{info: *b.info.Info}
	return info, nil
}

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
