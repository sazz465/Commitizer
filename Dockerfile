# FROM zenika/alpine-chrome:latest
# EXPOSE 9222
# CMD [chromium-browser, "--headless", "--disable-gpu", "--no-sandbox", "--remote-debugging-address=0.0.0.0", "--remote-debugging-port=9222"]





# Builder imahe to build go binary
FROM golang:latest
LABEL maintainer="iraj465 <saptarshimajumder19@gmail.com> "
RUN mkdir /app
ADD . /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o commitizer .
# CMD ["./commitizer"]