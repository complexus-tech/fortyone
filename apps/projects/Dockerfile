FROM node:22-alpine AS base

FROM base AS builder
RUN apk add --no-cache libc6-compat
RUN apk update

WORKDIR /app
RUN npm install pnpm turbo --global
RUN pnpm config set store-dir ~/.pnpm-store
COPY . .
RUN turbo prune projects --docker

FROM base AS installer
RUN apk add --no-cache libc6-compat
RUN apk update
WORKDIR /app

COPY .gitignore .gitignore
COPY --from=builder /app/out/json/ .
COPY --from=builder /app/out/pnpm-lock.yaml ./pnpm-lock.yaml
RUN pnpm install

# Build the project
COPY --from=builder /app/out/full/ .
# COPY turbo.json turbo.json

# Uncomment and use build args to enable remote caching
# ARG TURBO_TEAM
# ENV TURBO_TEAM=$TURBO_TEAM

# ARG TURBO_TOKEN
# ENV TURBO_TOKEN=$TURBO_TOKEN

RUN pnpm turbo build --filter=projects

FROM base AS runner
WORKDIR /app

# Don't run production as root
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
USER nextjs

COPY --from=installer /app/apps/projects/next.config.js .
COPY --from=installer /app/apps/projects/package.json .

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=installer --chown=nextjs:nodejs /app/apps/projects/.next/standalone ./
COPY --from=installer --chown=nextjs:nodejs /app/apps/projects/.next/static ./apps/projects/.next/static
COPY --from=installer --chown=nextjs:nodejs /app/apps/projects/public ./apps/projects/public

CMD node apps/projects/server.js