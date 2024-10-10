FROM golang:alpine

WORKDIR /app

COPY ./main.go ./
COPY ./go.mod ./

RUN go mod tidy
RUN go build

CMD ["/app/cts-times"]