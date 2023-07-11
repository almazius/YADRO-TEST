FROM golang:1.19-alpine
LABEL authors="sigy"

WORKDIR /yardo
COPY . .
RUN go build -o main api/cmd/main.go
CMD ["./main", "tests/test1.txt"]
