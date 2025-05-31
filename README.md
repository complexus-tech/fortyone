# Complexus.tech

A modern web platform built with Next.js, TypeScript, and Turborepo. The Complexus ecosystem consists of multiple interconnected applications served through local subdomains during development.

## Prerequisites

Before setting up the development environment, ensure you have the following installed:

### Required Tools

- **Node.js** (v18 or higher) - [Download here](https://nodejs.org/)
- **pnpm** (v9.3.0 or higher) - Install with `npm install -g pnpm`
- **Caddy** (v2 or higher) - [Installation guide](https://caddyserver.com/docs/install)

### Caddy Installation

Caddy is **required** for local development as it handles subdomain routing and SSL termination.

**macOS (Homebrew):**

```bash
brew install caddy
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

**Other platforms:** See [Caddy's official installation docs](https://caddyserver.com/docs/install)

## Local Development Setup

### 1. Clone and Install Dependencies

```bash
git clone <repository-url>
cd complexus.tech
pnpm install
```

### 2. Configure Local Domains

Add the following entries to your `/etc/hosts` file to enable subdomain routing:

```bash
sudo nano /etc/hosts
```

Add these lines:

```
127.0.0.1   complexus.local
127.0.0.1   docs.complexus.local
127.0.0.1   *.complexus.local
```

### 3. Start Development Environment

The project includes a unified development command that starts both Turbo and Caddy:

```bash
pnpm dev
```

This command will:

- Start all Next.js applications in development mode
- Launch Caddy server for subdomain routing
- Enable hot reloading across all apps

**Alternative commands:**

```bash
# Start only the Next.js apps (without Caddy)
pnpm dev:turbo

# Start only Caddy (if apps are running separately)
pnpm dev:caddy
```

### 4. Access Applications

Once running, access your applications at:

- **Landing Page**: http://complexus.local (port 3000)
- **Documentation**: http://docs.complexus.local (port 3002)
- **Projects App**: http://\*.complexus.local (port 3001)

## Code Structure

This is a Turborepo monorepo with the following structure:

```
complexus.tech/
├── apps/                    # Applications
│   ├── landing/            # Main landing page (complexus.local)
│   ├── docs/               # Documentation site (docs.complexus.local)
│   └── projects/           # Projects management app (*.complexus.local)
├── packages/               # Shared packages
│   ├── ui/                 # Shared React components
│   ├── icons/              # Icon library
│   ├── lib/                # Shared utilities and helpers
│   ├── tailwind-config/    # Shared Tailwind configuration
│   ├── eslint-config-custom/ # ESLint configuration
│   └── tsconfig/           # TypeScript configurations
├── Caddyfile              # Caddy server configuration
└── package.json           # Root package.json with scripts
```

### Applications

#### 🏠 Landing App (`apps/landing/`)

- **Purpose**: Main marketing and landing pages
- **URL**: http://complexus.local
- **Port**: 3000
- **Tech Stack**: Next.js 15, React 19, Framer Motion, GSAP
- **Features**:
  - MDX content support
  - Authentication with NextAuth
  - PostHog analytics
  - Cal.com integration

#### 📚 Docs App (`apps/docs/`)

- **Purpose**: Documentation and guides
- **URL**: http://docs.complexus.local
- **Port**: 3002
- **Tech Stack**: Next.js 15, Fumadocs
- **Features**:
  - MDX-based documentation
  - Built-in search
  - Code syntax highlighting

#### 🚀 Projects App (`apps/projects/`)

- **Purpose**: Main application for project management
- **URL**: http://\*.complexus.local (handles all subdomain routing)
- **Port**: 3001
- **Tech Stack**: Next.js 15, React 19, TanStack Query, Tiptap
- **Features**:
  - Rich text editing with Tiptap
  - Drag and drop functionality
  - Real-time collaboration
  - Jest testing setup
  - Docker support

### Shared Packages

#### 🎨 UI Package (`packages/ui/`)

Shared React component library built with:

- Radix UI primitives
- Tailwind CSS for styling
- TypeScript for type safety

#### 🔧 Lib Package (`packages/lib/`)

Shared utilities, helpers, and business logic used across applications.

#### 🎯 Icons Package (`packages/icons/`)

Centralized icon library for consistent iconography across all apps.

#### ⚙️ Configuration Packages

- **tailwind-config**: Shared Tailwind CSS configuration
- **eslint-config-custom**: Custom ESLint rules and configurations
- **tsconfig**: TypeScript configuration presets

## Development Workflow

### Building Applications

```bash
# Build all apps and packages
pnpm build

# Build specific app
pnpm build --filter=landing
pnpm build --filter=docs
pnpm build --filter=projects
```

### Linting and Formatting

```bash
# Lint all packages
pnpm lint

# Format code
pnpm format
```

### Testing

```bash
# Run tests (projects app has Jest setup)
cd apps/projects
pnpm test
```

## Networking Architecture

The development environment uses Caddy as a reverse proxy to route subdomain traffic:

- `complexus.local` → `localhost:3000` (landing)
- `docs.complexus.local` → `localhost:3002` (docs)
- `*.complexus.local` → `localhost:3001` (projects - wildcard routing)

This setup allows for:

- Realistic subdomain testing
- SSL/TLS termination in development
- Clean separation of applications
- Production-like routing behavior

## Troubleshooting

### Common Issues

**Subdomain not resolving:**

- Verify `/etc/hosts` entries are correct
- Clear DNS cache: `sudo dscacheutil -flushcache` (macOS)
- Restart Caddy: Stop with `Ctrl+C` and run `pnpm dev` again

**Port conflicts:**

- Check if ports 3000, 3001, 3002 are available
- Kill conflicting processes: `lsof -ti:3000 | xargs kill`

**Caddy not starting:**

- Verify Caddy installation: `caddy version`
- Check Caddyfile syntax: `caddy validate --config Caddyfile`
- Ensure no other web servers are running on port 80/443

**Dependencies issues:**

- Clear node_modules: `rm -rf node_modules && pnpm install`
- Clear Turbo cache: `pnpm turbo clean`
