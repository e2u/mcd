FROM alpine:latest

ENV  TZ="Asia/Shanghai"
RUN apk add --update --no-cache ca-certificates tzdata curl


RUN mkdir -p /logs/

COPY conf/ /opt/conf/
COPY docker/entrypoint.sh /opt/entrypoint.sh
COPY objs/mcd /opt/mcd


WORKDIR /opt/
EXPOSE 6000
EXPOSE 9000
ENTRYPOINT ["/opt/entrypoint.sh"]
