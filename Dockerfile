FROM golang

WORKDIR /root

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]