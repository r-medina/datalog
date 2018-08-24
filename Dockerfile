FROM alpine
COPY bin /go
RUN touch /var/log/access.log
EXPOSE 80/tcp
ENTRYPOINT [ "/go/datalog.linux-amd64" ]