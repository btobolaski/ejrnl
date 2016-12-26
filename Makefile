ejrnl: vendor/golang.org cmd/ejrnl crypto storage workflows ejrnl.go
	go test `glide nv`
	go build -o ejrnl code.tobolaski.com/brendan/ejrnl/cmd/ejrnl

glide.lock: glide.yaml
	glide update

vendor/golang.org: glide.lock
	glide install
