FROM clamav/clamav

RUN mkdir ~/uploads

RUN freshclam

ADD build/clamav-api /

CMD ["/clamav-api"]

EXPOSE 8080