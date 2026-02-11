# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

RUN apk add --no-cache python3 make g++

ARG NEXT_PUBLIC_WALLET_CONNECT_ID=demo
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ARG NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
ENV NEXT_PUBLIC_WALLET_CONNECT_ID=$NEXT_PUBLIC_WALLET_CONNECT_ID
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL

COPY package.json package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci --no-audit --no-fund; else npm install --no-audit --no-fund; fi

COPY . .
RUN npm run build

# Runtime stage
FROM node:20-alpine AS runner

WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/next.config.js ./next.config.js
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

CMD ["npm", "run", "start"]
