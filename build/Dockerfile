ARG GOLANG_VERSION=1.12.0
FROM golang:$GOLANG_VERSION as builder

WORKDIR /go/src/github.com/gradecak/watchdog

COPY . .
RUN go get github.com/Masterminds/glide; \
  glide install -v;

RUN build/build.sh;


FROM scratch

COPY --from=builder /go/src/github.com/gradecak/watchdog/watchdog /watchdog

ENTRYPOINT ["/watchdog"]

CMD [""]
