::build
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/ss5 cmd/server/main.go

SET GOOS=windows
SET GOARCH=amd64
go build -o bin/ss5.exe cmd/server/main.go

SET GOOS=linux
SET GOARCH=arm64
go build -o bin/ss5_arm cmd/server/main.go