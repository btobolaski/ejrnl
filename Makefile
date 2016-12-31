ejrnl: vendor/golang.org cmd/ejrnl crypto storage workflows ejrnl.go compression server
	go test `./glide nv`
	go build -o ejrnl github.com/btobolaski/ejrnl/cmd/ejrnl

glide.lock: glide glide.yaml
	./glide update

vendor/golang.org: glide glide.lock
	./glide install

vendor/github.com/Masterminds/glide:
	true

glide: glide.lock
	$(MAKE) -C vendor/github.com/Masterminds/glide build
	cp vendor/github.com/Masterminds/glide/glide ./
