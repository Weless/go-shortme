

mux:
go get -u github.com/gorilla/mux 

validator:
go get gopkg.in/validator.v2

log.LstdFlags会打印时间和日期，lost.Lshortfile会打印文件名及行号
```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```