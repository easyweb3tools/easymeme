FROM node:22-bookworm-slim

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

RUN npm install --no-audit --no-fund
RUN npm run build

ENV NODE_ENV=production

CMD ["sh", "-c", "if [ -f /app/openclaw.json ] && [ ! -f /tmp/openclaw-state/openclaw.json ]; then cp /app/openclaw.json /tmp/openclaw-state/openclaw.json; fi && ./node_modules/.bin/openclaw plugins install --link ./ && ./node_modules/.bin/openclaw plugins enable easymeme-openclaw-skill && while true; do ./node_modules/.bin/openclaw agent --local --session-id easymeme --message \"获取待分析代币 -> AI 分析 -> 回写结果\"; sleep 300; done"]
