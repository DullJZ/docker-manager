go env -w CGO_ENABLED=0 GOOS=windows GOARCH=amd64
go build -o ./build/docker-manager_windows_amd64.exe api.go funcs.go

go env -w CGO_ENABLED=0 GOOS=windows GOARCH=386
go build -o ./build/docker-manager_windows_386.exe api.go funcs.go

go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64
go build -o ./build/docker-manager_linux_amd64 api.go funcs.go

go env -w CGO_ENABLED=0 GOOS=linux GOARCH=386
go build -o ./build/docker-manager_linux_386 api.go funcs.go

go env -w CGO_ENABLED=0 GOOS=linux GOARCH=arm64
go build -o ./build/docker-manager_linux_arm64 api.go funcs.go

go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=amd64
go build -o ./build/docker-manager_darwin_amd64 api.go funcs.go

go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=arm64
go build -o ./build/docker-manager_darwin_arm64 api.go funcs.go
