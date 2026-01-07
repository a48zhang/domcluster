FROM golang:latest AS builder-ctl
WORKDIR /app
COPY . .
RUN make

FROM node:latest AS builder-web
WORKDIR /app
COPY --from=builder-ctl /app/web-ui .
COPY --from=builder-ctl /app/built ./built
RUN npm install && npm run build

FROM alpine:latest
WORKDIR /var/www/html
COPY --from=builder-web /app/built ./bin
COPY --from=builder-web /app/dist/* .
EXPOSE 80
RUN apk add --no-cache lighttpd
RUN printf 'server.document-root = "/var/www/html"\nserver.port = 80\nserver.indexfiles = ( "index.html", "index.htm" )\n' > /etc/lighttpd/lighttpd.conf
CMD ["lighttpd", "-f", "/etc/lighttpd/lighttpd.conf", "-D"]