FROM golang:1.23-alpine AS builder

WORKDIR /workspace

# Copy module metadata first for better layer caching.
COPY apps/easymeme/server/go.mod apps/easymeme/server/go.sum ./apps/easymeme/server/
COPY packages/go-sdk/go.mod ./packages/go-sdk/go.mod

# Build with a local workspace so easymeme can import easyweb3/go-sdk.
RUN go work init ./apps/easymeme/server ./packages/go-sdk
RUN cd apps/easymeme/server && go mod download

COPY apps/easymeme/server ./apps/easymeme/server
COPY packages/go-sdk ./packages/go-sdk
RUN cd apps/easymeme/server && CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=UTC

COPY --from=builder /bin/server /bin/server
EXPOSE 8080

CMD ["/bin/server"]
