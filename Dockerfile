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