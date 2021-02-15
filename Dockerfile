FROM golang:1.15.8
ADD . /go/src/github.com/locmai/blackping
WORKDIR /go/src/github.com/locmai/blackping
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o blackping .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/locmai/blackping/blackping .
CMD ["./blackping"]