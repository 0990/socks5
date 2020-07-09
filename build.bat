SET GOOS=windows
go build -o ./bin/ss5.exe ./cmd/server

SET GOOS=linux
go build -o ./bin/ss5 ./cmd/server