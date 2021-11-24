package ext2

import (
	"bytes"
	"encoding/binary"
	"io"
)

func extractLEStruct(data interface{},
					 readerFunc func([]byte) (int, error)) (size uint, err error){
	uSize := binary.Size(data)
	if err != nil {
		return
	}
	size = uint(uSize)

	bytesBuf := make([]byte, size)
	_, err = readerFunc(bytesBuf)
	if err != nil {
		return
	}

	err = binary.Read(bytes.NewReader(bytesBuf), binary.LittleEndian, data)
	return
}

func readLEStructAt(r io.ReaderAt, data interface{}, off int64) (size uint, err error){
	f := func(b []byte) (int, error) {
		return r.ReadAt(b, off)
	}
	return extractLEStruct(data, f)
}

func readLEStruct(r io.Reader, data interface{}) (size uint,err error){
	return extractLEStruct(data, r.Read)
}
