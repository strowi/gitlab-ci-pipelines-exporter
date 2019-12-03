dep:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -installsuffix cgo -o ./gitlab-ci-pipelines-exporter_linux_amd64
