package ext2

import (
	"io/fs"
	"os"
)

type FsImage struct {
	filename string
	file *os.File
	superBlock superBlock
	blockSize int64
}

func OpenImage (name string) (im *FsImage, err error) {
	im = new(FsImage)
	im.filename = name
	im.file, err = os.Open(name)

	_, err = readLEStructAt(im.file, &im.superBlock, 1024)
	if err != nil {
		return
	}

	im.blockSize = int64(1024) << im.superBlock.LogBlockSize
	return
}

func (im *FsImage) PrintFileOrDir(name string) (err error){
	file, err := im.open(name)
	if err != nil {
		return
	}
	switch file.mode() {
	case fTRegFile:
		return file.print()
	case fTDir:
		return file.printDir()
	default:
		return fs.ErrInvalid
	}
}

func (im *FsImage) Close() (err error) {
	err = im.file.Close()
	return
}
