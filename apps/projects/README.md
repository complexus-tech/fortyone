# FortyOne Projects App

[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![React](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.4-blue.svg)](https://www.typescriptlang.org/)
[![Jest](https://img.shields.io/badge/Jest-29-red.svg)](https://jestjs.io/)

The core FortyOne application for project management and team collaboration. Built with modern web technologies, it provides a comprehensive platform for organizing projects, tracking tasks, and collaborating with team members in real-time.

## ✨ Features

- **🎯 Project Management**: Create and manage projects with objectives and key results
- **📋 Task Tracking**: Organize tasks with rich text editing using Tiptap
- **👥 Team Collaboration**: Invite team members and collaborate in real-time
- **📝 Rich Text Editor**: Advanced text editing with formatting, links, and media
- **🎨 Drag & Drop**: Intuitive drag-and-drop interface for task management
- **📊 Analytics**: PostHog integration for usage insights
- **🔐 Workspace Security**: Subdomain-based workspace isolation
- **📱 Responsive Design**: Works seamlessly on desktop and mobile
- **⚡ Real-time Updates**: Live collaboration features
- **🔍 Search**: Powerful search across all content
- **📎 File Management**: Azure Blob Storage integration for file uploads
- **🧪 Testing**: Comprehensive Jest test suite

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- pnpm 9.3.0+
- Environment variables configured
- API backend running

### Development Setup

1. **Install dependencies**:

   ```bash
   cd apps/projects
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
   - Local: https://qa.fortyone.lc (requires Caddy for SSL)
   - Direct: http://localhost:3001

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

# AI Services
OPENAI_API_KEY=your_openai_api_key
GOOGLE_API_KEY=your_google_api_key
# Authentication
AUTH_SECRET=your_auth_secret_here
NEXTAUTH_SECRET=your_nextauth_secret

# Error Tracking
NEXT_PUBLIC_SENTRY_DSN=your_actual_sentry_dsn

# File Storage (Azure Blob Storage)
AZURE_STORAGE_ACCOUNT=your_azure_storage_account
AZURE_STORAGE_ACCESS_KEY=your_azure_storage_access_key
AZURE_STORAGE_CONTAINER=your_container_name
```

## 🧪 Testing

The projects app includes a comprehensive Jest test suite:

```bash
# Run all tests
pnpm test

# Run tests in watch mode
pnpm test:watch

# Run tests with coverage
pnpm test:coverage

# Run specific test file
pnpm test -- components/Button.test.tsx
```

### Test Structure

```
src/
├── components/          # Component tests
├── lib/                # Utility function tests
├── hooks/              # Custom hook tests
├── utils/              # Helper function tests
└── __tests__/          # Integration tests
```

## 🏗️ Architecture

### Tech Stack

- **Framework**: Next.js 15 with App Router
- **Runtime**: React 19 with concurrent features
- **State Management**: TanStack Query for server state
- **Styling**: Tailwind CSS with Radix UI components
- **Text Editor**: Tiptap for rich text editing
- **Authentication**: NextAuth.js
- **File Storage**: Azure Blob Storage
- **Error Tracking**: Sentry
- **Testing**: Jest + React Testing Library
- **Deployment**: Vercel or Docker

### Key Features

- **Workspace Routing**: Subdomain-based multi-tenancy
- **Real-time Collaboration**: Live updates and notifications
- **Rich Text Editing**: Full-featured editor with markdown support
- **File Uploads**: Secure file storage and management
- **API Integration**: RESTful API communication
- **Performance**: Optimized loading and caching

### Directory Structure

```
apps/projects/
├── app/                    # Next.js app router
│   ├── (workspace)/       # Protected workspace routes
│   ├── api/               # API routes and integrations
│   ├── login/             # Authentication pages
│   └── globals.css        # Global styles
├── components/            # React components
│   ├── shared/           # Shared UI components
│   ├── ui/               # Design system components
│   ├── modules/          # Feature modules
│   └── [feature]/        # Feature-specific components
├── lib/                  # Business logic
│   ├── actions/          # Server actions
│   ├── queries/          # API queries
│   ├── http/             # HTTP client utilities
│   └── utils/            # Helper functions
├── hooks/                # Custom React hooks
├── constants/            # App constants
├── types/                # TypeScript definitions
├── __tests__/            # Test files
└── utils/                # Utility functions
```

## 📦 Available Scripts

```bash
# Development
pnpm dev          # Start development server
pnpm build        # Build for production
pnpm start        # Start production server
pnpm lint         # Run ESLint
pnpm type-check   # Run TypeScript checking

# Testing
pnpm test         # Run Jest tests
pnpm test:watch   # Run tests in watch mode
pnpm test:coverage # Run tests with coverage report

# Database (if applicable)
pnpm db:migrate   # Run database migrations
pnpm db:seed      # Seed database with test data
```

## 🚀 Deployment

### Vercel (Recommended)

1. **Connect Repository**:

   - Import your GitHub repository to Vercel
   - Configure build settings:
     - Build Command: `cd apps/projects && pnpm build`
     - Output Directory: `apps/projects/.next`
     - Install Command: `pnpm install`

2. **Environment Variables**:

   - Add all environment variables to Vercel project settings
   - Configure subdomain routing for workspaces

3. **Custom Domains**:
   - Set up wildcard domain routing (\*.yourdomain.com)
   - Configure SSL certificates

```

## 🔧 Configuration

### Subdomain Routing

The app uses subdomain-based routing for workspace isolation:

- `qa.fortyone.lc` → QA workspace
- `marketing.fortyone.lc` → Marketing workspace
- `*.fortyone.lc` → Dynamic workspace routing

### File Storage

Azure Blob Storage is used for file uploads:

- **Security**: SAS tokens for secure access
- **Organization**: Files organized by workspace and user
- **Optimization**: Automatic image resizing and optimization

## 🤝 Contributing

See the main [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Workflow

1. **Pick an issue** from the GitHub issues
2. **Create a feature branch**: `git checkout -b feature/issue-number`
3. **Write tests** for new functionality
4. **Implement the feature**
5. **Run tests**: `pnpm test`
6. **Submit PR** with description and screenshots

## 📄 License

This project is licensed under the [FortyOne License](../../LICENSE).

## 📞 Support

- **Documentation**: [docs.fortyone.app](https://docs.fortyone.app)
- **Issues**: [GitHub Issues](https://github.com/complexus/fortyone/issues)
- **Discussions**: [GitHub Discussions](https://github.com/complexus/fortyone/discussions)
```
