#!/bin/sh
printf 'server.document-root = "/var/www/html"\nserver.port = 80\nserver.indexfiles = ( "index.html", "index.htm" )\n' > /etc/lighttpd/lighttpd.conf
./bin/d8rctl start
./bin/d8rctl password reset
lighttpd -f /etc/lighttpd/lighttpd.conf -D