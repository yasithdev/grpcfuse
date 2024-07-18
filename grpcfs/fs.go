// this file implements fuse -> grpc invocations

package grpcfs

import (
	"log"
	"os"
	"sync"

	"context"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

type grpcFs struct {
	fuseutil.NotImplementedFileSystem
	rootPath string
	inodes   *sync.Map
	logger   *log.Logger
}

var _ fuseutil.FileSystem = &grpcFs{}

// Create a file system that mirrors an existing physical path, in a readonly mode
func GrpcFsServer(rootPath string, logger *log.Logger) (server fuse.Server, err error) {

	if _, err = os.Stat(rootPath); err != nil {
		return nil, err
	}

	inodes := &sync.Map{}
	root := &inodeEntry{
		id:   fuseops.RootInodeID,
		path: rootPath,
	}
	inodes.Store(root.Id(), root)
	server = fuseutil.NewFileSystemServer(&grpcFs{
		rootPath: rootPath,
		inodes:   inodes,
		logger:   logger,
	})
	return
}

func (fs *grpcFs) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	return nil
}

func (fs *grpcFs) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	entry, err := getOrCreateInode(fs.inodes, op.Parent, op.Name)
	if err != nil {
		fs.logger.Printf("fs.LookUpInode for '%v' on '%v': %v", entry, op.Name, err)
		return fuse.EIO
	}
	if entry == nil {
		return fuse.ENOENT
	}
	outputEntry := &op.Entry
	outputEntry.Child = entry.Id()
	attributes, err := entry.Attributes()
	if err != nil {
		fs.logger.Printf("fs.LookUpInode.Attributes for '%v' on '%v': %v", entry, op.Name, err)
		return fuse.EIO
	}
	outputEntry.Attributes = *attributes
	return nil
}

func (fs *grpcFs) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	var entry, found = fs.inodes.Load(op.Inode)
	if !found {
		return fuse.ENOENT
	}
	attributes, err := entry.(Inode).Attributes()
	if err != nil {
		fs.logger.Printf("fs.GetInodeAttributes for '%v': %v", entry, err)
		return fuse.EIO
	}
	op.Attributes = *attributes
	return nil
}

func (fs *grpcFs) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	// Allow opening any directory.
	return nil
}

func (fs *grpcFs) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {
	var entry, found = fs.inodes.Load(op.Inode)
	if !found {
		return fuse.ENOENT
	}
	children, err := entry.(Inode).ListChildren(fs.inodes)
	if err != nil {
		fs.logger.Printf("fs.ReadDir for '%v': %v", entry, err)
		return fuse.EIO
	}

	if op.Offset > fuseops.DirOffset(len(children)) {
		return nil
	}

	children = children[op.Offset:]

	for _, child := range children {
		bytesWritten := fuseutil.WriteDirent(op.Dst[op.BytesRead:], *child)
		if bytesWritten == 0 {
			break
		}
		op.BytesRead += bytesWritten
	}
	return nil
}

func (fs *grpcFs) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) error {

	var _, found = fs.inodes.Load(op.Inode)
	if !found {
		return fuse.ENOENT
	}
	return nil

}

func (fs *grpcFs) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) error {
	var entry, found = fs.inodes.Load(op.Inode)
	if !found {
		return fuse.ENOENT
	}
	contents, err := entry.(Inode).Contents()
	if err != nil {
		fs.logger.Printf("fs.ReadFile for '%v': %v", entry, err)
		return fuse.EIO
	}

	if op.Offset > int64(len(contents)) {
		return fuse.EIO
	}

	contents = contents[op.Offset:]
	op.BytesRead = copy(op.Dst, contents)
	return nil
}

func (fs *grpcFs) ReleaseDirHandle(
	ctx context.Context,
	op *fuseops.ReleaseDirHandleOp) error {
	return nil
}

func (fs *grpcFs) GetXattr(
	ctx context.Context,
	op *fuseops.GetXattrOp) error {
	return nil
}

func (fs *grpcFs) ListXattr(
	ctx context.Context,
	op *fuseops.ListXattrOp) error {
	return nil
}

func (fs *grpcFs) ForgetInode(
	ctx context.Context,
	op *fuseops.ForgetInodeOp) error {
	return nil
}

func (fs *grpcFs) ReleaseFileHandle(
	ctx context.Context,
	op *fuseops.ReleaseFileHandleOp) error {
	return nil
}

func (fs *grpcFs) FlushFile(
	ctx context.Context,
	op *fuseops.FlushFileOp) error {
	return nil
}
