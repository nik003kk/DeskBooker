FROM node:22-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ARG REGISTRATION_EMAIL_DOMAIN=example.com
ENV VITE_API_BASE_URL=/api
ENV VITE_REGISTRATION_EMAIL_DOMAIN=$REGISTRATION_EMAIL_DOMAIN
RUN npm run build

FROM nginxinc/nginx-unprivileged:stable-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist/ /usr/share/nginx/html/
EXPOSE 8080
