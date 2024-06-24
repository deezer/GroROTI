FROM scratch

WORKDIR /

COPY groroti /
COPY config.toml /config.toml
COPY datadir/ /data

CMD [ "/groroti" ]
