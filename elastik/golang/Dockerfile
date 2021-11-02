FROM golang:latest

WORKDIR app
COPY . .
RUN export GIN_MODE=release
RUN go build main.go

CMD ["./main"]