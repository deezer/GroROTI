FROM scratch

WORKDIR /

COPY groroti /
COPY config.toml /config.toml

CMD [ "/groroti" ]
