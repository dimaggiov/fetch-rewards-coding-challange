FROM golang:1.19.3-alpine3.15
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go get -d github.com/gorilla/mux@latest
RUN go get -d github.com/google/uuid@latest
RUN go env -w GO111MODULE=on
RUN go build -o main . 
CMD ["/app/main"]
