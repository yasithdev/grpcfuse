// place for data-side helper functions

package grpcfs

import (
	"context"
	"fmt"
	"grpcfs/pb"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

var (
	uid                     = uint32(os.Getuid())
	gid                     = uint32(os.Getgid())
	allocatedInodeId uint64 = fuseops.RootInodeID
)

type Inode interface {
	Id() fuseops.InodeID
	Path() string
	String() string
	Attributes(ctx context.Context) (*fuseops.InodeAttributes, error)
	ListChildren(inodes *sync.Map, ctx context.Context) ([]*fuseutil.Dirent, error)
	Contents(ctx context.Context) ([]byte, error)
}

func getOrCreateInode(inodes *sync.Map, fsClient pb.GrpcFuseClient, ctx context.Context, parentId fuseops.InodeID, name string) (Inode, error) {
	parent, found := inodes.Load(parentId)
	if !found {
		return nil, nil
	}
	parentPath := parent.(Inode).Path()

	path := filepath.Join(parentPath, name)
	fileInfo, err := getStat(fsClient, ctx, path)
	if err != nil {
		return nil, nil
	}
	stat, _ := fileInfo.Sys().(*syscall.Stat_t)

	inodeEntry := &inodeEntry{
		id:   fuseops.InodeID(stat.Ino),
		path: path,
	}
	storedEntry, _ := inodes.LoadOrStore(inodeEntry.id, inodeEntry)
	return storedEntry.(Inode), nil
}

func nextInodeID() (next fuseops.InodeID) {
	nextInodeId := atomic.AddUint64(&allocatedInodeId, 1)
	return fuseops.InodeID(nextInodeId)
}

type inodeEntry struct {
	id     fuseops.InodeID
	path   string
	client pb.GrpcFuseClient
}

var _ Inode = &inodeEntry{}

func NewInode(path string, client pb.GrpcFuseClient) (Inode, error) {
	return &inodeEntry{
		id:     nextInodeID(),
		path:   path,
		client: client,
	}, nil
}

func (in *inodeEntry) Id() fuseops.InodeID {
	return in.id
}

func (in *inodeEntry) Path() string {
	return in.path
}

func (in *inodeEntry) String() string {
	return fmt.Sprintf("%v::%v", in.id, in.path)
}

func (in *inodeEntry) Attributes(ctx context.Context) (*fuseops.InodeAttributes, error) {
	fileInfo, err := getStat(in.client, ctx, in.path)
	if err != nil {
		return &fuseops.InodeAttributes{}, err
	}

	return &fuseops.InodeAttributes{
		Size:  uint64(fileInfo.Size()),
		Nlink: 1,
		Mode:  fileInfo.Mode(),
		Mtime: fileInfo.ModTime(),
		Uid:   uid,
		Gid:   gid,
	}, nil
}

func (in *inodeEntry) ListChildren(inodes *sync.Map, ctx context.Context) ([]*fuseutil.Dirent, error) {
	children, err := readDir(in.client, ctx, in.path)
	if err != nil {
		return nil, err
	}
	dirents := []*fuseutil.Dirent{}
	for i, child := range children {

		childInode, err := getOrCreateInode(inodes, in.client, ctx, in.id, child.Name())
		if err != nil || childInode == nil {
			continue
		}

		var childType fuseutil.DirentType
		if child.IsDir() {
			childType = fuseutil.DT_Directory
		} else if child.Type()&os.ModeSymlink != 0 {
			childType = fuseutil.DT_Link
		} else {
			childType = fuseutil.DT_File
		}

		dirents = append(dirents, &fuseutil.Dirent{
			Offset: fuseops.DirOffset(i + 1),
			Inode:  childInode.Id(),
			Name:   child.Name(),
			Type:   childType,
		})
	}
	return dirents, nil
}

func (in *inodeEntry) Contents(ctx context.Context) ([]byte, error) {
	res, err := readFile(in.client, ctx, in.path)
	return res, err
}
