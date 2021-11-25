# how to add an action to the api

1. declare it in `handlers.h`
2. define it in `handlers.cpp`
3. export it in `backend.go`
4. move to `pkg/backend/api.go` and implement what's needed from there
