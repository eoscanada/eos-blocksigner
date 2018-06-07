FROM gcr.io/cloud-builders/go as builder

ENV CGO_ENABLED 0
ENV PKG /root/go/src/github.com/eoscanada/eos-blocksigner
RUN mkdir -p $PKG
COPY . $PKG
RUN cd $PKG \
    && go test -v \
    && go build -v -o /eos-blocksigner

# Final image
FROM busybox
COPY --from=builder /eos-blocksigner /app/eos-blocksigner
