FROM golang:alpine as builder
RUN mkdir /build 
COPY . "/go/src/github.com/skwiwel/pod-contagion/"
WORKDIR "/go/src/github.com/skwiwel/pod-contagion/"
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /starter ./starter/main.go

FROM scratch
COPY --from=builder /starter .
ENTRYPOINT ["/starter"]