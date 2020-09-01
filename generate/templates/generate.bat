@ECHO OFF
oto -template server.go.plush^
	-out ../../pkg/avian-api/api.gen.go^
	-pkg api^
	../

gofmt -w ../../pkg/avian-api/api.gen.go ../../pkg/avian-api/api.gen.go

ECHO generated api.gen.go

@ECHO OFF
oto -template client.go.plush^
	-out ../../pkg/avian-client/avian.gen.go^
	-pkg avian^
	../
gofmt -w ../../pkg/avian-client/avian.gen.go ../../pkg/avian-client/avian.gen.go

ECHO generated avian.gen.go