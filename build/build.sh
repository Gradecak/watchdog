goos=${goos:-linux}
goarch=${goarch:-amd64}
# Build bundle
CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build\
           -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH}\
           -o "watchdog"\
           github.com/gradecak/watchdog/cmd/watchdog
