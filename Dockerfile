FROM golang:1.17.2-alpine

COPY . /app

WORKDIR /app

RUN go build -o /bin/tomato-demo

ENTRYPOINT ["/bin/tomato-demo"]