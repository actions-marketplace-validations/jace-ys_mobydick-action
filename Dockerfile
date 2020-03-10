FROM rust:1.41 AS builder
WORKDIR /usr/src/app
COPY . .
RUN rustup target add x86_64-unknown-linux-musl
RUN cargo install --target x86_64-unknown-linux-musl --path .

FROM alpine:3.11
COPY --from=builder /usr/local/cargo/bin/mobydick-action /usr/local/bin/mobydick-action
CMD ["mobydick-action"]
