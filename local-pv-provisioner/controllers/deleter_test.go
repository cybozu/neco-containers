package controllers

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testFillDeleter() {
	var Ki int64 = 1024
	var Mi int64 = 1024 * Ki
	var Gi int64 = 1024 * Mi

	It("should fill first specified bytes with zero", func() {
		tmpFile, _ := os.CreateTemp("", "deleter")
		defer os.Remove(tmpFile.Name())
		err := exec.Command("dd", `if=/dev/urandom`, "of="+tmpFile.Name(), "bs=1M", "count=1025").Run()
		Expect(err).ShouldNot(HaveOccurred())

		deleter := &FillDeleter{
			FillBlockSize: 1024,
			FillCount:     10,
		}
		deleter.Delete(tmpFile.Name())

		zeroBlock := make([]byte, deleter.FillBlockSize)
		buffer := make([]byte, deleter.FillBlockSize)
		for _, position := range []int64{0, 1 * Gi} {
			tmpFile.Seek(position, 0)
			for i := uint(0); i < deleter.FillCount; i++ {
				_, err := tmpFile.Read(buffer)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cmp.Equal(buffer, zeroBlock)).Should(BeTrue())
			}
		}

		_, err = tmpFile.Read(buffer)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(cmp.Equal(buffer, zeroBlock)).Should(BeFalse())
	})

	fillBlockSize := uint(1024)
	fillCount := uint(10)
	DescribeTable(
		"ensure that all bdev labels created by Ceph are zapped",
		func(size int64, expectedZappedRanges []contiguousRange) {
			useTestFillDeleterFS(func(testFS *testFillDeleterFS) {
				fileName := "test-file"

				err := testFS.fallocate(fileName, size)
				Expect(err).NotTo(HaveOccurred())

				deleter := &FillDeleter{
					FillBlockSize: fillBlockSize,
					FillCount:     fillCount,
				}
				err = deleter.Delete(fileName)
				Expect(err).NotTo(HaveOccurred())

				zappedRanges, err := testFS.getWrittenRanges(fileName)
				Expect(err).NotTo(HaveOccurred())
				Expect(zappedRanges).Should(Equal(expectedZappedRanges))
			})
		},
		Entry("file size < 1Gi", 500*Mi, []contiguousRange{
			{from: 0, to: 10 * Ki},
		}),
		Entry("file size = 1Gi", 1*Gi, []contiguousRange{
			{from: 0, to: 10 * Ki},
		}),
		Entry("1Gi < file size < 1Gi+10Ki", 1*Gi+5*Ki, []contiguousRange{
			{from: 0, to: 10 * Ki},
			{from: 1 * Gi, to: 1*Gi + 5*Ki},
		}),
		Entry("file size = 1Gi+10Ki", 1*Gi+10*Ki, []contiguousRange{
			{from: 0, to: 10 * Ki},
			{from: 1 * Gi, to: 1*Gi + 10*Ki},
		}),
		Entry("1Gi+10Ki < file size < 10Gi", 2*Gi, []contiguousRange{
			{from: 0, to: 10 * Ki},
			{from: 1 * Gi, to: 1*Gi + 10*Ki},
		}),
		Entry("file size = 10Ti", 10*1024*Gi, []contiguousRange{
			{from: 0, to: 10 * Ki},
			{from: 1 * Gi, to: 1*Gi + 10*Ki},
			{from: 10 * Gi, to: 10*Gi + 10*Ki},
			{from: 100 * Gi, to: 100*Gi + 10*Ki},
			{from: 1000 * Gi, to: 1000*Gi + 10*Ki},
		}),
	)
}

type contiguousRange struct {
	from, to int64 // [from, to)
}

type testFillDeleterFSFile struct {
	writtenRanges []contiguousRange
	name          string
	offset, size  int64
}

func (file *testFillDeleterFSFile) Close() error {
	return nil
}

func (file *testFillDeleterFSFile) Write(p []byte) (n int, err error) {
	to := file.offset + int64(len(p))

	if file.size < to {
		return 0, errors.New("p is too large")
	}

	if len(file.writtenRanges) != 0 && file.writtenRanges[len(file.writtenRanges)-1].to == file.offset {
		file.writtenRanges[len(file.writtenRanges)-1].to = to
	} else {
		file.writtenRanges = append(file.writtenRanges, contiguousRange{
			from: file.offset,
			to:   to,
		})
	}

	file.offset = to

	return len(p), nil
}

func (file *testFillDeleterFSFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		file.offset = offset
	case io.SeekEnd:
		if offset != 0 {
			return 0, errors.New("not implemented")
		}
		file.offset = file.size
	default:
		return 0, errors.New("not implemented")
	}
	return file.offset, nil
}

func (file *testFillDeleterFSFile) Stat() (FileInfo, error) {
	return nil, errors.New("not implemented")
}

func (file *testFillDeleterFSFile) Sync() error {
	return nil
}

type testFillDeleterFS struct {
	files map[string]*testFillDeleterFSFile
	notImplementedFS
}

var _ fileSystem = &testFillDeleterFS{}

func NewTestFillDeleterFS() *testFillDeleterFS {
	return &testFillDeleterFS{
		files: map[string]*testFillDeleterFSFile{},
	}
}

func (fs *testFillDeleterFS) fallocate(name string, size int64) error {
	_, ok := fs.files[name]
	if ok {
		return errors.New("already exists")
	}
	fs.files[name] = &testFillDeleterFSFile{
		name:          name,
		size:          size,
		writtenRanges: []contiguousRange{},
	}
	return nil
}

func (fs *testFillDeleterFS) getWrittenRanges(name string) ([]contiguousRange, error) {
	file, ok := fs.files[name]
	if !ok {
		return nil, errors.New("file not found")
	}
	return file.writtenRanges, nil
}

func (fs *testFillDeleterFS) OpenFile(name string, flag int, perm FileMode) (file, error) {
	file, ok := fs.files[name]
	if !ok {
		return nil, errors.New("file not found")
	}
	return file, nil
}

func useTestFillDeleterFS(f func(testFS *testFillDeleterFS)) {
	originalFS := fs
	testFS := NewTestFillDeleterFS()
	fs = testFS
	defer func() { fs = originalFS }()
	f(testFS)
}
