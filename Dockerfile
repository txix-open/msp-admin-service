# Version: 0.0.1
FROM alpine:latest

#RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -v

RUN mkdir app

RUN apk add --no-cache tzdata
RUN cp /usr/share/zoneinfo/UTC /etc/localtime
RUN echo "UTC" >  /etc/timezone

ADD admin-service /app/admin-service
ADD conf/config.yml /app/config.yml
ADD static /app/static

ENTRYPOINT ["/app/admin-service"]

EXPOSE 5000
