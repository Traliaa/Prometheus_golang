FROM golang:latest

WORKDIR app
COPY . .

RUN export export ELASTIC_APM_SERVER_URL=http://apm-server:8200
RUN go build main.go

CMD ["./main"]