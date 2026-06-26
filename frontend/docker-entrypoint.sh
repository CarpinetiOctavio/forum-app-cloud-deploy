#!/bin/sh
set -e
API_URL="${REACT_APP_API_URL:-http://localhost:8080/api}"
find /usr/share/nginx/html -name '*.js' -exec sed -i "s|__REACT_APP_API_URL__|${API_URL}|g" {} \;
exec nginx -g 'daemon off;'
