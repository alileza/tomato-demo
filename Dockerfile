FROM golang

COPY . /app

WORKDIR /app

RUN go build -o /bin/tomato-demo

ENTRYPOINT ["/bin/tomato-demo"]