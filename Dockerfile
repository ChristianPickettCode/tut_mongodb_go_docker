# grab base image
FROM golang:1.19

WORKDIR /go/src/app

COPY ./src .

RUN go get -d -v

RUN go build -v

CMD ["./tut_mongodb_go_docker"]
