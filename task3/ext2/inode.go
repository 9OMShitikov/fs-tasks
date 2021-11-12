package ext2

import (
	"os"
	"strings"
	"unsafe"
)

type groupDescriptor struct {
	BlockBitmap			uint32
	INodeBitmap			uint32
	INodeTable			uint32
	FreeBlocksCount		uint16
	FreeInodesCount		uint16
	UsedDirsCount		uint16
	Flags				uint16
	Reserved			[3]uint32
};

type iNodeRecord struct {
	Mode		uint16
	Uid			uint16
	Size		uint32
	ATime		uint32
	CTime		uint32
	MTime		uint32
	DTime		uint32
	Gid			uint16
	LinksCount	uint16
	Blocks		uint32
	Flags		uint32
	Osd1		uint32
	Block		[nBlocks]uint32
	Generation	uint32
	FileAcl		uint32
	DirAcl		uint32
	FAddr		uint32
	Osd2		[12]byte
}

func (im *FsImage) getFile(iNodeId int64) (file *fsFile, err error) {
	file = new(fsFile)
	file.im = im
	file.pos = 0
	file.blockId = [4]int{0, 0, 0, 0}
	file.iNodeId = iNodeId
	file.depth = 0

	blockGroup := (iNodeId - 1) / int64(im.superBlock.InodePerGroup)
	index := (iNodeId - 1) % int64(im.superBlock.InodePerGroup)
	descSize := int64(unsafe.Sizeof(file.desc))
	groupDescAddr := im.blockSize + descSize * blockGroup

	_, err = readLEStructAt(im.file, &file.desc, groupDescAddr)
	if err != nil {
		return
	}
	inodePos := int64(file.desc.INodeTable) * im.blockSize + index * int64(im.superBlock.INodeSize)
	_, err = readLEStructAt(im.file, &file.iNode, inodePos)

	file.size = int64(file.iNode.Size)
	if file.isReg() {
		file.size |= int64(file.iNode.DirAcl) << 32
	}
	return
}

func (im *FsImage) open(name string) (file *fsFile, err error) {
	parts := strings.Split(name, "/")

	inode := int64(rootIno)
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		file, err = im.getFile(inode)
		if err != nil {
			return
		}

		var dirContents []directoryEntry
		dirContents, err = file.readDir()
		if err != nil {
			return
		}

		found := false
		for _, entry := range dirContents {
			if entry.Name == part {
				found = true
				inode = int64(entry.Header.Inode)
			}
		}

		if !found {
			return nil, os.ErrNotExist
		}
	}

	file, err = im.getFile(inode)
	return
}

