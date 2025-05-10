FROM golang:1.24.2 as builder
WORKDIR /usr/src/app
COPY . .
WORKDIR /usr/src/app/book-recomendation
RUN go mod download

EXPOSE 8080

CMD ["go", "run", "/usr/src/app/book-recomendation/cmd/server", "."]