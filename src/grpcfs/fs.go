// place for fuse definitions

package grpcfs

import (
	"context"
	"log"
	"sync"

	pb "grpcfs/pb"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcFs struct {
	fuseutil.NotImplementedFileSystem
	root   string
	inodes *sync.Map
	logger *log.Logger
	client pb.FuseServiceClient
}

var _ fuseutil.FileSystem = &grpcFs{}

// Create a file system that mirrors an existing physical path, in a readonly mode
func FuseServer(
	grpcHost string,
	root string,
	logger *log.Logger) (server fuse.Server, err error) {

	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	client := pb.NewFuseServiceClient(conn)

	if _, err = getStat(client, context.TODO(), root); err != nil {
		logger.Print("error in getStat() for FS root", err)
		return nil, err
	}

	inodes := &sync.Map{}
	rootInode := &inodeEntry{
		id:     fuseops.RootInodeID,
		path:   root,
		client: client,
	}
	inodes.Store(rootInode.Id(), rootInode)
	server = fuseutil.NewFileSystemServer(&grpcFs{
		root:   root,
		inodes: inodes,
		logger: logger,
		client: client,
	})
	return
}

func (fs *grpcFs) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {

	res, err := getStatFs(fs.client, ctx, fs.root)
	if err != nil {
		return err
	}
	op.BlockSize = res.BlockSize
	op.Blocks = res.Blocks
	op.BlocksAvailable = res.BlocksAvailable
	op.BlocksFree = res.BlocksFree
	op.Inodes = res.Inodes
	op.InodesFree = res.InodesFree
	op.IoSize = res.IoSize
	return nil
}

func (fs *grpcFs) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	fs.logger.Print("fs.LookUpInode - called. ", op)
	entry, err := getOrCreateInode(fs.inodes, fs.client, ctx, op.Parent, op.Name)
	if err == nil && entry == nil {
		fs.logger.Print("fs.LookUpInode - file does not exist. ", op.Name)
		return fuse.ENOENT
	}
	if err != nil {
		fs.logger.Printf("fs.LookUpInode - '%v' on '%v': %v", entry, op.Name, err)
		return fuse.EIO
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
	log.Print("fs.ReadDir - called. ", op.Inode)
	var entry, found = fs.inodes.Load(op.Inode)
	if !found {
		log.Print("fs.ReadDir - requested dir not found. ", op.Inode)
		return fuse.ENOENT
	}
	log.Print("fs.ReadDir - found requested dir. ", entry)
	children, err := entry.(Inode).ListChildren(fs.inodes)
	log.Print("fs.ReadDir - requested children. ", entry)
	if err != nil {
		fs.logger.Printf("fs.ReadDir - ListChildren of '%v' failed: %v", entry, err)
		return fuse.EIO
	}
	fs.logger.Printf("fs.ReadDir - Got children of '%v': %v", entry, children)
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
