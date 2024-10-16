package storage

type Store interface {
	// Key-Value Getters
	GetString(key string) string
	GetStringWithDefault(key string, value string) string
	GetBoolean(key string) bool
	GetBooleanWithDefault(key string, value bool) bool
	GetInt(key string) int
	GetIntWithDefault(key string, value int) int
	GetLong(key string) int64
	GetLongWithDefault(key string, value int64) int64
	GetByteArray(key string) []byte
	GetByteArrayWithDefault(key string, value []byte) []byte

	// Key-Value Setters
	SetString(key string, value string)
	SetBoolean(key string, value bool)
	SetInt(key string, value int)
	SetLong(key string, value int64)
	SetByteArray(key string, value []byte)

	Open() error
	Close() error
	Contains(key string) (bool, error)
	Remove(key string) error
	Clear() error
}
