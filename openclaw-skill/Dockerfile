FROM node:22-bookworm

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
COPY ./openclaw /app/openclaw
COPY entrypoint.sh /app/entrypoint.sh

RUN npm install --no-audit --no-fund
RUN npm run build
RUN npm install -g openclaw

ENV NODE_ENV=production
ENV HOME=/home/node

RUN chown -R node:node /app
USER node

CMD ["sh", "/app/entrypoint.sh"]
