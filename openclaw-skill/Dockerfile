FROM node:22-bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates git python3 make g++ cmake \
    && rm -rf /var/lib/apt/lists/*

RUN git config --global url."https://github.com/".insteadOf "ssh://git@github.com/"

COPY package.json package-lock.json* tsconfig.json .
COPY src ./src
COPY SKILL.md ./SKILL.md
COPY openclaw.plugin.json ./openclaw.plugin.json
COPY skills ./skills

RUN npm install --no-audit --no-fund
RUN npm run build
RUN npm install -g openclaw@latest

ENV NODE_ENV=production

CMD ["sh", "-c", "openclaw plugins install --link ./ && openclaw plugins enable easymeme-openclaw-skill && while true; do openclaw agent --local --session-id easymeme --message \"获取待分析代币 -> AI 分析 -> 回写结果\"; sleep 300; done"]
