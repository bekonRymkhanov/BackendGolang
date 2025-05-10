FROM golang:1.23.0 as builder
WORKDIR /usr/src/app
COPY . .
WORKDIR /usr/src/app/book-service
RUN go mod download

EXPOSE 4000

CMD ["go", "run", "/usr/src/app/book-service/cmd/server", "."]