FROM gcr.io/cloud-builders/go as builder

ENV CGO_ENABLED 0
ENV PKG /root/go/src/github.com/eoscanada/eos-kms-block-signer
RUN mkdir -p $PKG
COPY . $PKG
RUN cd $PKG \
    && go get -v -t ./eos-kms-block-signer

RUN cd $PKG/eos-kms-block-signer \
    && go test -v \
    && go build -v  -o /eos-kms-block-signer

FROM busybox

COPY --from=builder /eos-kms-block-signer /app/eos-kms-block-signer
