::build
SET GOOS=linux
go build -o bin/ss5 cmd/server/main.go
::go build -o bin/udp_test cmd/client/udp_test/main.go

SET GOOS=windows
go build -o bin/ss5.exe cmd/server/main.go
::go build -o bin/udp_test.exe cmd/client/udp_test/main.go