# FortyOne

[![License](https://img.shields.io/badge/license-FortyOne%20License-blue.svg)](LICENSE)
[![Node.js Version](https://img.shields.io/badge/node-%3E%3D18.0.0-brightgreen)](https://nodejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.4-blue.svg)](https://www.typescriptlang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)

A modern, open-source web platform for project management and collaboration. FortyOne provides a comprehensive suite of tools for teams to organize, track, and deliver projects efficiently.

## ✨ Features

- **🎯 Project Management**: Organize projects with objectives, key results, and milestones
- **📋 Task Tracking**: Create, assign, and track tasks with rich text editing
- **👥 Team Collaboration**: Real-time collaboration with team members
- **📊 Analytics**: PostHog integration for user analytics and insights
- **🔐 Authentication**: Secure authentication with NextAuth and Google OAuth
- **📱 Cross-Platform**: Web application with React Native mobile app
- **🎨 Modern UI**: Built with Radix UI, Tailwind CSS, and Framer Motion
- **⚡ Performance**: Optimized with Next.js 15 and React 19
- **🔍 Search**: Built-in search functionality across all content
- **📚 Documentation**: Comprehensive documentation site with Fumadocs

## 📸 Screenshots

### Core Features

#### 📋 List View - Story Management

![List View](apps/landing/public/images/product/list.webp)

#### 📊 Kanban Board - Task Workflow

![Kanban Board](apps/landing/public/images/product/kanban-light.webp)

#### 🎯 OKR Objectives - Goal Tracking

![OKR Objectives](apps/landing/public/images/product/objective.webp)

#### 🤖 Maya AI Assistant

![Maya AI Assistant](apps/landing/public/images/product/maya.webp)

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/complexus/fortyone.git
cd fortyone

# Install dependencies
pnpm install

# Set up environment variables
cp apps/landing/.env.example apps/landing/.env
cp apps/projects/.env.example apps/projects/.env
cp apps/mobile/.env.example apps/mobile/.env
# Edit .env files with your actual values

# Start development server
pnpm dev
```

Visit [https://fortyone.lc](https://fortyone.lc) to see the application.

## 📖 Documentation

- **[📚 User Guide](https://docs.fortyone.app)** - Complete documentation
- **[🤝 Contributing](CONTRIBUTING.md)** - How to contribute
- **[🔒 Security](SECURITY.md)** - Security policy and reporting

## 🏗️ Architecture

FortyOne is built as a monorepo using Turborepo with the following structure:

```
fortyone/
├── apps/                    # Applications
│   ├── landing/            # Marketing & authentication
│   ├── docs/               # Documentation site
│   ├── projects/           # Main project management app
│   └── mobile/             # React Native mobile app
├── packages/               # Shared packages
│   ├── ui/                 # Component library
│   ├── lib/                # Shared utilities
│   ├── icons/              # Icon library
│   └── [config]/           # Tooling configurations
└── tools/                  # Development tools
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development

```bash
# Install dependencies
pnpm install

# Start development
pnpm dev

# Run tests
pnpm test

# Lint code
pnpm lint
```

## 📄 License

This project is licensed under the [FortyOne License](LICENSE) - see the [LICENSE](LICENSE) file for details.

The FortyOne License allows free personal and non-commercial use, while requiring commercial licensing for businesses with $200K+ annual revenue.

## 📞 Contact & Support

- **📧 Email**: hello@complexus.tech
- **🐛 Issues**: [GitHub Issues](https://github.com/complexus/fortyone/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/complexus/fortyone/discussions)
- **📖 Documentation**: [docs.fortyone.app](https://docs.fortyone.app)

## 🙏 Acknowledgments

- Built with [Next.js](https://nextjs.org/), [React](https://reactjs.org/), and [TypeScript](https://www.typescriptlang.org/)
- UI components by [Radix UI](https://www.radix-ui.com/)
- Analytics by [PostHog](https://posthog.com/)

---

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
cd fortyone
pnpm install
```

### 2. Configure Local Domains

**Important:** We use `.lc` domains for local development instead of `.local` because `.local` domains cannot be added to OAuth providers (like Google) as valid redirect URIs. This allows us to test authentication flows locally with real OAuth providers.

#### Domain Strategy

- **`.local`**: Cannot be used for OAuth redirects, but Caddy supports automatic wildcard routing
- **`.localhost`**: Caddy supports automatic wildcard routing, but limited OAuth provider support
- **`.lc`**: Valid TLD for OAuth redirects, but requires manual subdomain configuration

Add the following entries to your `/etc/hosts` file:

```bash
sudo nano /etc/hosts
```

Add these lines:

```
127.0.0.1   fortyone.lc
127.0.0.1   docs.fortyone.lc
127.0.0.1   www.fortyone.lc
127.0.0.1   qa.fortyone.lc
127.0.0.1   payments.fortyone.lc
127.0.0.1   growth.fortyone.lc
```

**Note:** Unlike `.local` or `.localhost` domains, each subdomain must be manually added to both `/etc/hosts` and the `Caddyfile`. When adding new workspaces or subdomains, you'll need to:

1. Add the subdomain to `/etc/hosts`
2. Add a corresponding entry in the `Caddyfile`
3. Restart Caddy

### 3. Start Development Environment

The project includes a unified development command that starts both Turbo and Caddy:

```bash
pnpm dev
```

This command will:

- Start all Next.js applications in development mode
- Launch Caddy server with SSL/TLS termination
- Enable hot reloading across all apps
- Serve apps with HTTPS (using Caddy's internal CA)

### 4. Access Applications

Once running, access your applications at:

- **Landing Page**: https://fortyone.lc (port 3000)
- **Documentation**: https://docs.fortyone.lc (port 3002)
- **Projects App**: https://\*.fortyone.lc (port 3001)
  - QA workspace: https://qa.fortyone.lc
  - Payments workspace: https://payments.fortyone.lc
  - Growth workspace: https://growth.fortyone.lc

**Note:** All local domains use HTTPS with Caddy's internal CA. You may need to accept the security certificate in your browser on first visit.

## Code Structure

This is a Turborepo monorepo with the following structure:

```
fortyone.tech/
├── apps/                    # Applications
│   ├── landing/            # Main landing page (fortyone.lc)
│   ├── docs/               # Documentation site (docs.fortyone.lc)
│   └── projects/           # Projects management app (*.fortyone.lc)
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
- **URL**: https://fortyone.lc
- **Port**: 3000
- **Tech Stack**: Next.js 15, React 19, Framer Motion, GSAP
- **Features**:
  - MDX content support
  - Authentication with NextAuth
  - PostHog analytics
  - Cal.com integration

#### 📚 Docs App (`apps/docs/`)

- **Purpose**: Documentation and guides
- **URL**: https://docs.fortyone.lc
- **Port**: 3002
- **Tech Stack**: Next.js 15, Fumadocs
- **Features**:
  - MDX-based documentation
  - Built-in search
  - Code syntax highlighting

#### 🚀 Projects App (`apps/projects/`)

- **Purpose**: Main application for project management
- **URL**: https://\*.fortyone.lc (handles workspace-specific subdomain routing)
- **Port**: 3001
- **Tech Stack**: Next.js 16, React 19, TanStack Query, Tiptap
- **Features**:
  - Rich text editing with Tiptap
  - Drag and drop functionality
  - Real-time collaboration
  - Workspace-based subdomain routing
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

### Adding New Subdomains

When adding a new workspace or subdomain:

1. **Add to `/etc/hosts`:**

   ```bash
   sudo nano /etc/hosts
   # Add: 127.0.0.1   newworkspace.fortyone.lc
   ```

2. **Update `Caddyfile`:**

   ```caddy
   newworkspace.fortyone.lc {
       tls internal
       reverse_proxy localhost:3001
   }
   ```

3. **Restart Caddy:**
   ```bash
   # Stop current dev process (Ctrl+C) and restart
   pnpm dev
   ```

## Networking Architecture

The development environment uses Caddy as a reverse proxy to route subdomain traffic:

- `fortyone.lc` → `localhost:3000` (landing)
- `docs.fortyone.lc` → `localhost:3002` (docs)
- `*.fortyone.lc` → `localhost:3001` (projects - workspace-specific routing)

This setup allows for:

- **OAuth compatibility**: `.lc` domains work with Google OAuth and other providers
- **SSL/TLS in development**: Caddy provides HTTPS with internal certificates
- **Realistic subdomain testing**: Production-like routing behavior
- **Clean application separation**: Each app runs independently

### Why .lc domains?

- **OAuth Support**: Unlike `.local`, `.lc` domains can be registered as valid redirect URIs with OAuth providers like Google, enabling full authentication testing in development
- **Valid TLD**: `.lc` is Saint Lucia's country code TLD, making it a legitimate domain for authentication services
- **Manual Control**: While requiring manual configuration, this gives explicit control over which subdomains are available

## Troubleshooting

### Common Issues

**Subdomain not resolving:**

- Verify `/etc/hosts` entries are correct
- Ensure the subdomain is added to the `Caddyfile`
- Clear DNS cache: `sudo dscacheutil -flushcache` (macOS) or `sudo systemctl restart systemd-resolved` (Linux)
- Restart Caddy: Stop with `Ctrl+C` and run `pnpm dev` again

**SSL Certificate warnings:**

- Accept Caddy's internal CA certificate in your browser
- The warning is expected on first visit to each subdomain
- Consider installing Caddy's root certificate for seamless development

**Port conflicts:**

- Check if ports 3000, 3001, 3002 are available
- Kill conflicting processes: `lsof -ti:3000 | xargs kill`

**Caddy not starting:**

- Verify Caddy installation: `caddy version`
- Check Caddyfile syntax: `caddy validate --config Caddyfile`
- Ensure no other web servers are running on port 80/443

**OAuth authentication issues:**

- Verify the subdomain is registered in your OAuth provider's console
- Ensure redirect URIs include the full `https://subdomain.fortyone.lc` format
- Check that the subdomain resolves correctly before testing auth

**Dependencies issues:**

- Clear node_modules: `rm -rf node_modules && pnpm install`
- Clear Turbo cache: `pnpm turbo clean`

- credits to dominic