#
# STAGE 1: build static web files
#
FROM node:14 as frontend
WORKDIR /src

#
# install dependencies
COPY client/package*.json ./
RUN npm install

#
# build client
COPY client/ .
RUN npm run build

#
# STAGE 2: build executable binary
#
FROM golang:1.16-buster as builder
WORKDIR /app

COPY . .
RUN go get -v -t -d .; \
    go build -o bin/go4tv cmd/go4tv/main.go

#
# STAGE 3: build a small image
#
#FROM scratch
#COPY --from=builder /app/bin/go4tv /app/bin/go4tv
COPY --from=frontend /src/dist/ /var/www

ENV GO4TV_BIND=:8080
ENV GO4TV_STATIC=/var/www

EXPOSE 8080

ENTRYPOINT [ "/app/bin/go4tv" ]
CMD [ "serve" ]
