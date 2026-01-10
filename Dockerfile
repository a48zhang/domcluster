FROM golang:alpine AS builder-ctl
WORKDIR /app
COPY . .
RUN mkdir -p built && \
    cd d8rctl && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o ../built/d8rctl main.go && cd .. && \
    cd domclusterd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o ../built/domclusterd main.go && cd .. && \
    cp docker-entry.sh built/docker-entry.sh

FROM node:latest AS builder-web
WORKDIR /app
COPY ./web-ui .
COPY --from=builder-ctl /app/built ./built
RUN npm install && npm run build

FROM alpine:latest
WORKDIR /var/www/html
COPY --from=builder-web /app/built ./bin
COPY --from=builder-web /app/dist/* .
RUN chmod +x ./bin/docker-entry.sh && apk add --no-cache lighttpd && apk add --no-cache libc6-compat gcompat
EXPOSE 80 50051 18080
CMD ["./bin/docker-entry.sh"]