FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-metabase"]
COPY baton-metabase /