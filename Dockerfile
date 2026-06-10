FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY groroti /
COPY config.toml.default /config.toml
COPY --chown=1000:2000 datadir/ /data

EXPOSE 3000

# Run as an unprivileged user by default, matching the Helm chart's
# runAsUser:1000 / fsGroup:2000. distroless/static also ships CA
# certificates and tzdata, which the scratch base lacked.
USER 1000:2000

CMD [ "/groroti" ]
