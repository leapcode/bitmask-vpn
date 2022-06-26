package openvpn

type ErrorFromServer []byte

func (err ErrorFromServer) Error() string {
	return string(err)
}

func (err ErrorFromServer) String() string {
	return string(err)
}
