package ext2

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
)

type directoryEntryHeader struct {
	Inode uint32
	RecLen uint16
	NameLen uint8
	Flags uint8
}

type directoryEntry struct {
	Header directoryEntryHeader
	Name string
}

type fsFile struct {
	iNodeId int64
	iNode iNodeRecord
	desc groupDescriptor
	im *FsImage

	size int64
	pos int64
	blockOffset int64
	blockNum uint32

	// Indexes of blocks for every layer
	blockId [4]int
	depth int
}

func (file *fsFile) nextBlock() (err error) {
	file.blockNum += 1
	if file.blockNum == file.iNode.Blocks {
		return io.ErrUnexpectedEOF
	}
	changedLevel := 0
	if file.depth == 0 {
		file.blockId[0] += 1
		if file.blockId[0] >= indBlock {
			file.depth += 1
			changedLevel = 0
		}
	} else {
		maxBlockPointers := int(file.im.blockSize / 4)
		changedLevel = file.depth
		for ;changedLevel != 0 && file.blockId[changedLevel] == maxBlockPointers - 1; changedLevel -= 1 {
			file.blockId[changedLevel] = 0
		}
		file.blockId[changedLevel] += 1
		if changedLevel == 0 && file.blockId[0] == nBlocks {
			return io.ErrUnexpectedEOF
		}
		if changedLevel == 0 {
			file.depth += 1
		}
	}
	return
}

func (file *fsFile) getBlockPtr() (off int64, err error) {
	ind := file.iNode.Block[file.blockId[0]]
	for i := 1; i <= file.depth; i++ {
		addr := int64(ind) * file.im.blockSize + int64(4 * file.blockId[i])
		_, err = readLEStructAt(file.im.file, &ind, addr)
		if err != nil {
				return
			}
	}
	return int64(ind) * file.im.blockSize, nil
}

func (file *fsFile) Read(b []byte) (n int, err error) {
	size := int64(len(b))
	if file.pos == file.size {
		return 0, io.EOF
	}
	if file.pos + size > file.size {
		size = file.size - file.pos
	}
	start := int64(0)

	for size > 0 {
		var ptr int64
		ptr, err = file.getBlockPtr()
		if err != nil {
			return
		}
		readSize := file.im.blockSize - file.blockOffset
		if readSize > size {
			readSize = size
		}
		_, err = file.im.file.ReadAt(b[start: start + readSize],
									 ptr + file.blockOffset)
		if err != nil {
			return 0, err
		}
		start += readSize
		size -= readSize
		if readSize + file.blockOffset == file.im.blockSize {
			file.blockOffset = 0
			err = file.nextBlock()
			if err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					return int(start), err
				}
			}
		} else {
			file.blockOffset += readSize
		}
	}
	return int(start), nil
}

func (file *fsFile) mode () uint8 {
	return uint8(file.iNode.Mode>>12)
}

func (file *fsFile) isDir() bool {
	return file.iNode.Mode>>12 == fTDir
}

func (file *fsFile) isReg() bool {
	return file.mode() == fTRegFile
}

func (file *fsFile) readDir() (dirs []directoryEntry, err error) {
	if !file.isDir() {
		err = fs.ErrInvalid
		return
	}

	rem := file.size
	for rem > 0 {
		var entry directoryEntry
		var headerSize uint
		headerSize, err = readLEStruct(file, &entry.Header)
		if err != nil {
			return
		}
		if entry.Header.RecLen < 9 {
			err = fs.ErrInvalid
			return
		}

		nameBytes := make([]byte, uint(entry.Header.RecLen) - headerSize)
		_, err = file.Read(nameBytes)
		if err != nil {
			return
		}
		entry.Name = string(nameBytes[:entry.Header.NameLen])
		rem -= int64(entry.Header.RecLen)
		dirs = append(dirs, entry)
	}
	return
}

func (file *fsFile) print() (err error) {
	const bufLen = 4096
	var buf [bufLen]byte

	size := file.size
	var readSize int
	for i := 0; i < 20 && size > 0; i++ {
		readSize, err = file.Read(buf[:])
		if err != nil {
			return
		}
		size -= int64(readSize)
		fmt.Print(string(buf[:readSize]))
	}
	return
}

func (file *fsFile) printDir() (err error) {
	if !file.isDir() {
		err = fs.ErrInvalid
		return
	}

	res, err := file.readDir()
	if err != nil {
		return
	}

	for _, dir := range res {
		fmt.Println(dir.Name)
	}
	return
}
