ejrnl: vendor/golang.org
	go test `glide nv`

glide.lock: glide.yaml
	glide update

vendor/golang.org: glide.lock
	glide install
