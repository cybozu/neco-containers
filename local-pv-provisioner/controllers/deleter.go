package controllers

// Deleter clean up the block device
type Deleter interface {
	Delete(path string) error
}

// FillDeleter fills first 100MByte with '\0'
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

	zeroBlock := make([]byte, d.FillBlockSize)
	for i := uint(0); i < d.FillCount; i++ {
		_, err = file.Write(zeroBlock)
		if err != nil {
			return err
		}
	}
	return file.Sync()
}
