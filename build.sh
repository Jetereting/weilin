set GOOS=linux
set GOARCH=amd64
go build -ldflags "-w -s" -i -o weilin

#or
GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -i -o weilin