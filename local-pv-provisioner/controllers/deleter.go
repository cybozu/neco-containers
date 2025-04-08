package controllers

import (
	"fmt"
	"io"
)

// Deleter clean up the block device
type Deleter interface {
	Delete(path string) error
}

type FillDeleter struct {
	FillBlockSize uint
	FillCount     uint
}

// Delete implements Deleter's method.
func (d *FillDeleter) Delete(path string) error {
	file, err := fs.OpenFile(path, O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// cf. https://github.com/ceph/ceph/blob/v19.2.1/src/os/bluestore/BlueStore.cc#L138-L143
	bdevLabelPositions := []int64{
		0,
		1 * 1024 * 1024 * 1024,
		10 * 1024 * 1024 * 1024,
		100 * 1024 * 1024 * 1024,
		1000 * 1024 * 1024 * 1024,
	}

	// Get the size of the device. Note that we can't use file.Stat() here,
	// because its Size() always returns 0 for a device file.
	fileSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to get device size: %w", err)
	}

	zeroBlock := make([]byte, d.FillBlockSize)
	for _, position := range bdevLabelPositions {
		if position >= fileSize {
			break
		}
		_, err := file.Seek(position, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek: %d: %w", position, err)
		}
		for i := uint(0); i < d.FillCount; i++ {
			from := position + int64(i)*int64(len(zeroBlock))
			to := position + int64(i+1)*int64(len(zeroBlock))
			length := len(zeroBlock)
			if fileSize <= from {
				break
			} else if fileSize <= to {
				length = int(fileSize - from)
			}
			_, err = file.Write(zeroBlock[0:length])
			if err != nil {
				return err
			}
		}
	}
	return file.Sync()
}
