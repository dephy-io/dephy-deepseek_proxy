FROM rust:1.84 AS builder
WORKDIR /app

COPY Cargo.toml Cargo.lock ./

COPY dsproxy-controller ./dsproxy-controller
COPY balance-payment-sdk ./balance-payment-sdk

RUN cargo build --release --bin dephy-dsproxy-controller

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/target/release/dephy-dsproxy-controller .
CMD ["./dephy-dsproxy-controller", "--nostr-relay", "ws://host.docker.internal:8000", "--machine-pubkeys", "d041ea9854f2117b82452457c4e6d6593a96524027cd4032d2f40046deb78d93", "--admin-pubkey", "d041ea9854f2117b82452457c4e6d6593a96524027cd4032d2f40046deb78d93", "--solana-rpc-url", "http://host.docker.internal:8899"]