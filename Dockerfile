# Commitizer
FROM golang:alpine
LABEL maintainer="iraj465 <saptarshimajumder19@gmail.com> "
RUN mkdir /app
ADD . /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o commitizer .

# Headless chrome
FROM zenika/alpine-chrome:latest
CMD ["chromium-browser", "--headless", "--no-sandbox", "--remote-debugging-port=9222"]