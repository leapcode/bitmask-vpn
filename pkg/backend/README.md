# how to add new fields to backend api

1. add to the struct in `status.go`
2. populate it in `toJson()` method.
3. modify the `bitmask` interface in `pkg/bitmask/bitmask.go` (note: this is
   a relict from the past, we can probably get rid of since there'll be
   a single implementation in the foreseeable future).
4. modify the `bitmask` struct in `pkg/vpn/main.go`
5. modify the bitmask instantiation in `pkg/vpn/main.go:Init`
6. implement functionality...
