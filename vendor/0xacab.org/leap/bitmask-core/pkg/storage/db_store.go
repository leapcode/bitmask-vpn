package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asdine/storm/v3"
)

type DBStore struct {
	Store
	db       *storm.DB
	filepath string
}

const defaultBucket = "defaultBucket"

func NewDBStore(path string) (*DBStore, error) {
	fp := filepath.Join(path, "bitmask.db")
	err := os.MkdirAll(filepath.Dir(fp), 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create DB store path: %v", err)
	}
	db, err := storm.Open(fp)

	if err != nil {
		return nil, fmt.Errorf("failed to open storm DB: %v", err)
	}

	return &DBStore{db: db, filepath: fp}, nil
}

func (s *DBStore) GetString(key string) string {
	var value string
	err := s.db.Get(defaultBucket, key, &value)
	if err != nil {
		return ""
	}
	return value
}

func (s *DBStore) GetStringWithDefault(key string, value string) string {
	var result string
	err := s.db.Get(defaultBucket, key, &value)
	if err != nil {
		return value
	}
	return result
}

func (s *DBStore) GetBoolean(key string) bool {
	var value bool
	err := s.db.Get(defaultBucket, key, &value)
	if err != nil {
		return false
	}
	return value
}

func (s *DBStore) GetBooleanWithDefault(key string, value bool) bool {
	var result bool
	err := s.db.Get(defaultBucket, key, &result)
	if err != nil {
		return value
	}
	return result
}

func (s *DBStore) GetInt(key string) int {
	var value int
	err := s.db.Get(defaultBucket, key, &value)
	if err != nil {
		return 0
	}
	return value
}
func (s *DBStore) GetIntWithDefault(key string, value int) int {
	var result int
	err := s.db.Get(defaultBucket, key, &result)
	if err != nil {
		return value
	}
	return result
}
func (s *DBStore) GetLong(key string) int64 {
	var value int64
	err := s.db.Get(defaultBucket, key, &value)
	if err != nil {
		return 0
	}
	return value
}

func (s *DBStore) GetLongWithDefault(key string, value int64) int64 {
	var result int64
	err := s.db.Get(defaultBucket, key, &result)
	if err != nil {
		return value
	}
	return result
}

func (s *DBStore) GetByteArray(key string) []byte {
	var value []byte
	err := s.db.Get(defaultBucket, key, &value)
	if err != nil {
		return nil
	}
	return value
}

func (s *DBStore) GetByteArrayWithDefault(key string, value []byte) []byte {
	var result []byte
	err := s.db.Get(defaultBucket, key, &result)
	if err != nil {
		return value
	}
	return result
}

// Key-Value Setters
func (s *DBStore) SetString(key string, value string) {
	_ = s.db.Set(defaultBucket, key, value)
}
func (s *DBStore) SetBoolean(key string, value bool) {
	_ = s.db.Set(defaultBucket, key, value)
}

func (s *DBStore) SetInt(key string, value int) {
	_ = s.db.Set(defaultBucket, key, value)
}
func (s *DBStore) SetLong(key string, value int64) {
	_ = s.db.Set(defaultBucket, key, value)
}
func (s *DBStore) SetByteArray(key string, value []byte) {
	_ = s.db.Set(defaultBucket, key, value)
}

func (s *DBStore) Contains(key string) (bool, error) {
	return s.db.KeyExists(defaultBucket, key)
}
func (s *DBStore) Remove(key string) error {
	return s.db.Delete(defaultBucket, key)
}
func (s *DBStore) Clear() error {
	return s.db.Drop(defaultBucket)
}

func (s *DBStore) Close() error {
	return s.db.Close()
}

func (s *DBStore) Open() error {
	db, err := storm.Open(s.filepath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	s.db = db
	return nil
}
