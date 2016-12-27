ejrnl: vendor/golang.org cmd/ejrnl crypto storage workflows ejrnl.go
	go test `./glide nv`
	go build -o ejrnl code.tobolaski.com/brendan/ejrnl/cmd/ejrnl

glide.lock: glide glide.yaml
	./glide update

vendor/golang.org: glide glide.lock
	./glide install

vendor/github.com/Masterminds/glide:
	true

glide: glide.lock
	$(MAKE) -C vendor/github.com/Masterminds/glide build
	cp vendor/github.com/Masterminds/glide/glide ./
