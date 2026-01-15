# AGENTS.md

## Repository Overview

- Monorepo managed with `pnpm` workspaces and Turborepo.
- Main apps live in `apps/` and shared packages in `packages/`.
- Tooling configs are centralized in `packages/eslint-config-custom` and `packages/tsconfig`.

## Key Paths

- `apps/landing`: Next.js 16 marketing site.
- `apps/docs`: Next.js 15 docs site (Fumadocs).
- `apps/projects`: Next.js 16 main app (Jest tests live here).
- `apps/server`: Go backend API (Projects API).
- `apps/mobile`: Expo/React Native app.
- `packages/ui`: shared React component library (Tailwind styles).
- `packages/lib`: shared utilities.
- `packages/icons`: icon library.
- `packages/tailwind-config`: shared Tailwind CSS config.
- `packages/eslint-config-custom`: ESLint configs (Vercel style guide).
- `packages/tsconfig`: TypeScript base configs.

## Environment & Setup

- Node.js `>=18` and `pnpm@9.3.0`.
- Go `>=1.23` for `apps/server` (see `apps/server/README.md`).
- Docker & Docker Compose are required for the server stack.
- `air` is used for live reload in `apps/server` (`go install github.com/air-verse/air@latest`).
- Caddy is required for full local dev routing (see `README.md`).
- Install dependencies: `pnpm install`.

## Build / Lint / Test Commands

### Root (Turborepo)

- Build everything: `pnpm build` (runs `turbo run build`).
- Lint everything: `pnpm lint` (runs `turbo run lint`).
- Format code: `pnpm format` (Prettier for `ts/tsx/md`).
- Dev + Caddy: `pnpm dev`.
- Dev without Caddy: `pnpm dev:caddy` (runs Turbo only).

### App-Specific Commands

- Landing dev: `pnpm --filter landing dev`.
- Landing build: `pnpm --filter landing build`.
- Landing lint: `pnpm --filter landing lint`.
- Docs dev: `pnpm --filter docs dev`.
- Docs build: `pnpm --filter docs build`.
- Docs start: `pnpm --filter docs start`.
- Projects dev: `pnpm --filter projects dev`.
- Projects build: `pnpm --filter projects build`.
- Projects lint: `pnpm --filter projects lint`.
- Projects lint (fix): `pnpm --filter projects lint:fix`.
- Projects type check: `pnpm --filter projects type-check`.
- Mobile start: `pnpm --filter mobile start`.
- Mobile lint: `pnpm --filter mobile lint`.
- Mobile iOS: `pnpm --filter mobile ios`.
- Mobile Android: `pnpm --filter mobile android`.

### Package Commands

- UI build styles: `pnpm --filter ui build:styles`.
- UI build components: `pnpm --filter ui build:components`.
- UI dev styles: `pnpm --filter ui dev:styles`.
- UI dev components: `pnpm --filter ui dev:components`.
- UI lint: `pnpm --filter ui lint`.
- UI type check: `pnpm --filter ui check-types`.
- Icons lint: `pnpm --filter icons lint`.

### Tests (Jest is only configured in projects)

- Run all tests: `pnpm --filter projects test`.
- Run a single test file: `pnpm --filter projects test -- path/to/test.test.ts`.
- Run a test name pattern: `pnpm --filter projects test -- -t "pattern"`.
- From `apps/projects`: `pnpm test` or `pnpm test -- -t "pattern"`.

### Server (Go API)

- Dev server (air): `make dev` (run in `apps/server`).
- Worker process: `make worker`.
- Run API without live reload: `make develop`.
- Go module tidy: `make tidy`.
- Start Jaeger (OTel): `make otel`.
- Stop Jaeger: `make stop-otel`.
- Create migration: `make migrate-create name=add_table_name`.
- Apply migrations: `make migrate-up`.
- Rollback migrations: `make migrate-down n=1`.
- Check migration version: `make migrate-version`.
- Force migration version: `make migrate-force v=2`.
- Seed data: `make seed` (supports `name`, `slug`, `email`, `fullname`).

## Formatting

- Prettier is the formatter (`pnpm format`).
- Tailwind class sorting is enabled via `prettier-plugin-tailwindcss`.
- Use the existing formatting (double quotes, semicolons, trailing commas).
- Go code should be formatted with `gofmt`.

## TypeScript & ESLint

- Base config: `packages/tsconfig/base.json` (strict mode enabled).
- Next.js apps extend `packages/tsconfig/nextjs.json`.
- ESLint extends the Vercel Engineering Style Guide.
- Import resolver uses `tsconfig.json` in each workspace.
- `apps/projects` has `@/*` alias for `src/*`.

### TypeScript Defaults

- `strict: true` and `strictNullChecks: true` across shared configs.
- `moduleResolution: "Bundler"` and `module: "ESNext"` in base config.
- `allowImportingTsExtensions: true` for explicit `.ts`/`.tsx` imports.
- `noEmit: true` in shared configs and Next.js apps.
- `esModuleInterop: true` for interop-friendly imports.

### ESLint Config Locations

- Next.js apps: `packages/eslint-config-custom/next.js`.
- React libraries: `packages/eslint-config-custom/react-internal.js`.
- Node/TS libraries: `packages/eslint-config-custom/library.js`.
- `import/no-default-export` is disabled in custom configs.

### Local ESLint Overrides (Apps)

- Projects app relaxes several strict TS rules (no-unsafe-\*, require-await).
- Projects app allows `react/no-array-index-key` and `import/no-cycle`.
- Landing app allows `@next/next/no-img-element` and disables some strict TS rules.
- Both apps disable `jsx-a11y/no-autofocus` and `react/function-component-definition`.

## Imports

- Prefer explicit `import type` for type-only imports.
- Group imports: types first, third-party next, local relative last.
- Keep local imports relative to the file or use workspace aliases.

## React / Next.js Conventions

- Use functional components and React hooks.
- Add `"use client"` only when needed for client-side hooks.
- Prefer named exports; default exports are allowed when needed.
- Keep server components free of browser-only APIs.

## Naming Conventions

- Components: `PascalCase` filenames and exports.
- Hooks: `useX` prefix.
- Variables/functions: `camelCase`.
- Constants: `UPPER_SNAKE_CASE` for module-level constants.

## Error Handling & Safety

- Favor early returns and clear error messages.
- Keep API/async errors surfaced to the UI or logging tools already in use.
- Avoid swallowing errors; use `try/catch` only when recovery is explicit.

## Testing Conventions

- Jest + `@testing-library/react` in `apps/projects`.
- Add new tests alongside related modules when possible.
- Use `jest.setup.ts` for test environment setup.

## Cursor / Copilot Rules

- No `.cursor/rules`, `.cursorrules`, or `.github/copilot-instructions.md` found.
