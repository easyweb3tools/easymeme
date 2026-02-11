FROM node:22-bookworm AS builder

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git python3 make g++ cmake \
    && rm -rf /var/lib/apt/lists/*

RUN git config --global url."https://github.com/".insteadOf "ssh://git@github.com/"

COPY package.json package-lock.json* tsconfig.json .
COPY index.ts ./index.ts
COPY src ./src
COPY skills ./skills
COPY openclaw.plugin.json ./openclaw.plugin.json
COPY entrypoint.sh /app/entrypoint.sh

RUN npm ci --no-audit --no-fund
RUN npm run build

FROM node:22-bookworm-slim AS runtime

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g openclaw --no-audit --no-fund

COPY package.json package-lock.json* ./
RUN npm ci --omit=dev --no-audit --no-fund

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/skills ./skills
COPY --from=builder /app/openclaw.plugin.json ./openclaw.plugin.json
COPY --from=builder /app/index.ts ./index.ts
COPY --from=builder /app/entrypoint.sh /app/entrypoint.sh

ENV NODE_ENV=production
ENV HOME=/home/node

RUN chown -R node:node /app
USER node

CMD ["sh", "/app/entrypoint.sh"]
