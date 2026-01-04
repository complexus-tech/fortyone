# FortyOne Landing App

[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.4-blue.svg)](https://www.typescriptlang.org/)

The FortyOne landing page and marketing website, built with Next.js 15, React 19, and modern web technologies. This app serves as the main entry point for users, providing authentication, marketing content, and access to the FortyOne platform.

## ✨ Features

- **🏠 Marketing Pages**: Homepage, features, pricing, contact, and blog
- **🔐 Authentication**: User registration and login with NextAuth
- **📧 Email Integration**: Magic link authentication and notifications
- **📊 Analytics**: PostHog integration for user tracking
- **🎨 Modern UI**: Framer Motion animations and GSAP interactions
- **📱 Responsive Design**: Mobile-first approach with Tailwind CSS
- **📝 MDX Content**: Blog posts and documentation pages
- **🔍 SEO Optimized**: Meta tags, structured data, and performance optimized
- **🌐 Multi-domain**: Supports subdomain routing for different workspaces

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- pnpm 9.3.0+
- Environment variables configured

### Development Setup

1. **Install dependencies**:

   ```bash
   cd apps/landing
   pnpm install
   ```

2. **Set up environment variables**:

   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

3. **Start development server**:

   ```bash
   pnpm dev
   ```

4. **Access the application**:
   - Local: https://fortyone.lc (requires Caddy for SSL)
   - Direct: http://localhost:3000

### Environment Variables

Create a `.env` file with the following variables:

```bash
# Domain Configuration
NEXT_PUBLIC_DOMAIN=fortyone.lc

# API Configuration
NEXT_PUBLIC_API_URL=api_url

# Analytics
NEXT_PUBLIC_POSTHOG_KEY=your_posthog_key
NEXT_PUBLIC_POSTHOG_HOST=https://app.posthog.com
NEXT_PUBLIC_GOOGLE_ANALYTICS_ID=your_google_analytics_id

# AI Services
OPENAI_API_KEY=your_openai_api_key
GOOGLE_API_KEY=your_google_api_key

# Authentication
AUTH_SECRET=your_auth_secret_here
NEXTAUTH_SECRET=your_nextauth_secret

# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret

# Error Tracking (Optional)
NEXT_PUBLIC_SENTRY_DSN=your_sentry_dsn
```

## 🏗️ Architecture

### Tech Stack

- **Framework**: Next.js 15 with App Router
- **Runtime**: React 19 with concurrent features
- **Styling**: Tailwind CSS with custom design system
- **Animations**: Framer Motion + GSAP
- **Content**: MDX for blog posts and documentation
- **Authentication**: NextAuth.js with Google OAuth
- **Analytics**: PostHog for user tracking
- **Deployment**: Vercel or any Node.js hosting

### Key Components

- **Authentication Flow**: Login/register with magic links and Google OAuth
- **Workspace Routing**: Subdomain-based routing for different workspaces
- **Content Management**: MDX-powered blog and documentation
- **SEO Optimization**: Dynamic meta tags and structured data
- **Performance**: Image optimization, lazy loading, and caching

### Directory Structure

```
apps/landing/
├── app/                    # Next.js app router
│   ├── (marketing)/       # Public marketing pages
│   ├── (onboarding)/      # User onboarding flow
│   ├── api/               # API routes
│   └── globals.css        # Global styles
├── components/            # React components
│   ├── shared/           # Reusable UI components
│   ├── ui/               # Design system components
│   └── [feature]/        # Feature-specific components
├── content/              # MDX content
├── lib/                  # Business logic and utilities
├── hooks/                # Custom React hooks
├── constants/            # App constants
├── types/                # TypeScript type definitions
└── utils/                # Helper functions
```

## 📦 Available Scripts

```bash
# Development
pnpm dev          # Start development server
pnpm build        # Build for production
pnpm start        # Start production server
pnpm lint         # Run ESLint
pnpm type-check   # Run TypeScript type checking

# Content
pnpm postinstall  # Generate MDX content (runs automatically)
```

## 🚀 Deployment

### Vercel (Recommended)

1. **Connect Repository**:

   - Import your GitHub repository to Vercel
   - Configure build settings:
     - Build Command: `cd apps/landing && pnpm build`
     - Output Directory: `apps/landing/.next`
     - Install Command: `pnpm install`

2. **Environment Variables**:

   - Add all environment variables from `.env` to Vercel project settings

3. **Custom Domain**:
   - Configure `fortyone.app` as the production domain
   - Set up subdomain routing for workspaces

### Other Platforms

The app can be deployed to any platform supporting Node.js:

- **Railway**
- **Render**
- **Fly.io**
- **Self-hosted** with Docker

## 🤝 Contributing

See the main [Contributing Guide](../../CONTRIBUTING.md) for details on:

- Setting up the development environment
- Code style and standards
- Submitting pull requests
- Reporting issues

## 📄 License

This project is licensed under the [FortyOne License](../../LICENSE).

## 📞 Support

- **Documentation**: [docs.fortyone.app](https://docs.fortyone.app)
- **Issues**: [GitHub Issues](https://github.com/complexus/fortyone/issues)
- **Discussions**: [GitHub Discussions](https://github.com/complexus/fortyone/discussions)
