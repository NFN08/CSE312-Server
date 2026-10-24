FROM golang
# double check this new command
RUN apt-get update && apt-get install -y ffmpeg

WORKDIR /root

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]