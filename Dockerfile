FROM scratch

WORKDIR /

COPY groroti /
COPY config.toml.default /config.toml
COPY datadir/ /data

CMD [ "/groroti" ]
