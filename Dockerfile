FROM clamav/clamav

RUN mkdir ~/uploads
RUN mkdir run/lock

RUN freshclam

ADD build/clamav-api /

CMD ["/clamav-api"]

EXPOSE 8080