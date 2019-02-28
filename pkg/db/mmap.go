package db

import (
	"os"
	"syscall"
)

// MMapDb ...
// Refs: https://medium.com/@arpith/adventures-with-mmap-463b33405223
type MMapDb struct {
	filename string
	Data     []byte
	fd       int
	file     *os.File
}

// NewMMapDb ...
func NewMMapDb(filename string) *MMapDb {
	return &MMapDb{filename: filename}
}

// Mmap ...
func (db *MMapDb) Mmap(size int) error {
	data, err := syscall.Mmap(db.fd, 0, size, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	db.Data = data
	return nil
}

// Resize ...
func (db *MMapDb) Resize(size int) error {
	err := syscall.Ftruncate(db.fd, int64(size))
	if err != nil {
		return err
	}
	return nil
}

// Open ...
func (db *MMapDb) Open() error {
	f, err := os.OpenFile(db.filename, os.O_CREATE|os.O_RDWR, 0)
	if err != nil {
		return err
	}
	db.fd = int(f.Fd())
	db.file = f
	return nil
}
