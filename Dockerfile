FROM golang:alpine
LABEL authors="vorticist"
WORKDIR /app
COPY . .
RUN go build -o the-account .
CMD ["/app/the-account"]