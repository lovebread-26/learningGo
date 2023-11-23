FROM golang:1.20.6 as build
WORKDIR /build
ENV GOPROXY https://goproxy.cn
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o gw .

FROM alpine:latest AS final
WORKDIR /app
COPY --from=build  /build/gw /app/
COPY --from=build  /build/blacklist.json /app/


CMD ["./gw", "-dockerStyle", "true"]