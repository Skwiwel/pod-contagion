FROM golang:alpine as builder
RUN mkdir /build 
COPY . "/go/src/github.com/skwiwel/pod-contagion/"
WORKDIR "/go/src/github.com/skwiwel/pod-contagion/"
RUN ls
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /podder app/podder/main.go

FROM scratch
COPY --from=builder /podder .
WORKDIR /app
CMD ["/podder"]