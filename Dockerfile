FROM alpine AS builder

RUN apk update && apk add git go

RUN git clone https://github.com/zbblanton/cloudflare_dynamic_dns.git && \
    cd cloudflare_dynamic_dns && \
    go build

FROM alpine

COPY --from=builder cloudflare_dynamic_dns/cloudflare_dynamic_dns /app/

