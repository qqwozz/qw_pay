FROM debian:bookworm-slim AS builder
RUN apt-get update && apt-get install -y g++ libhiredis-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY fraud_engine.cpp Makefile ./
RUN make

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libhiredis1.1 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/fraud_engine .
CMD ["./fraud_engine"]
