# Build from the repository root:
# docker build -f apps/collaboration/Dockerfile -t fortyone-collaboration .
FROM node:22.16-alpine AS dependencies
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.3.0 --activate
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/collaboration/package.json apps/collaboration/package.json
COPY packages/document-editor/package.json packages/document-editor/package.json
RUN pnpm install --filter collaboration... --prod --frozen-lockfile --ignore-scripts

FROM node:22.16-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=dependencies /app /app
COPY apps/collaboration/src apps/collaboration/src
COPY apps/collaboration/tsconfig.json apps/collaboration/tsconfig.json
COPY packages/document-editor/src packages/document-editor/src
USER node
EXPOSE 1234
CMD ["node", "--import", "./apps/collaboration/node_modules/tsx/dist/loader.mjs", "apps/collaboration/src/index.ts"]
