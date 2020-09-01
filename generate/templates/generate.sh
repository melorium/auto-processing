oto -template server.go.plush \
	-out ../pkg/api/server.gen.go \
	-pkg api \
	../pkg/api/def
gofmt -w server.gen.go server.gen.go
echo "generated server.gen.go"

oto -template client.go.plush \
	-out ../pkg/api/client.gen.go \
	-pkg api \
	../pkg/api/def
gofmt -w client.gen.go client.gen.go
echo "generated client.gen.go"