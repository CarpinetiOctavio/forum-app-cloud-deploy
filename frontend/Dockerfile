FROM node:24.19.0-alpine3.24 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --ignore-scripts
COPY . .
ARG REACT_APP_API_URL=__REACT_APP_API_URL__
RUN npm run build

FROM nginx:1.31.3-alpine3.24
RUN sed -i 's#root   /usr/share/nginx/html;#root /tmp/html;#' \
      /etc/nginx/conf.d/default.conf \
 && sed -i 's#include /etc/nginx/conf.d/\*.conf;#include /tmp/conf.d/*.conf;#' \
      /etc/nginx/nginx.conf \
 && chown -R nginx:nginx /var/cache/nginx /run
COPY --from=builder /app/build /usr/share/nginx/html
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
USER nginx
EXPOSE 8080
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
