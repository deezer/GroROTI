FROM scratch

WORKDIR /

RUN mkdir /data && chown 1000: /data

COPY groroti /
COPY config.toml /config.toml

CMD [ "/groroti" ]
