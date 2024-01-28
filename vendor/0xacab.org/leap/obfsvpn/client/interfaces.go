package client

type ObfsClient interface {
	Start() (bool, error)
	Stop() (bool, error)
	IsStarted() bool
}
