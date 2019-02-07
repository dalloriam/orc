FROM golang:1.11
WORKDIR /go/src/github.com/dalloriam/orc/

COPY . .
RUN make static

FROM alpine:latest  
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /app/
COPY --from=0 /go/src/github.com/dalloriam/orc/dist/orc .

ENTRYPOINT [ "/app/orc" ]

CMD [ "server" ]