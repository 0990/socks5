::build
SET GOOS=linux
go build -o bin/ss5 cmd/server/main.go

SET GOOS=windows
go build -o bin/ss5.exe cmd/server/main.go

::release
::https://goreleaser.com/quick-start/
::goreleaser --snapshot --skip-publish --rm-dist
::goreleaser --rm-dist