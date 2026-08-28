# FortyOne Documentation

[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![Fumadocs](https://img.shields.io/badge/Fumadocs-15-blue.svg)](https://fumadocs.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.4-blue.svg)](https://www.typescriptlang.org/)

The customer documentation site for FortyOne, built with Next.js and Fumadocs. It provides user guides, API documentation, developer resources, and product support content for the managed FortyOne service.

## ✨ Features

- **📚 Comprehensive Documentation**: User guides, API docs, and tutorials
- **🔍 Full-Text Search**: Fast, client-side search across all content
- **📝 MDX Support**: Rich content with code blocks, tables, and components
- **🎨 Modern UI**: Clean, responsive design with dark/light mode
- **🚀 Fast Performance**: Optimized loading and navigation
- **📱 Mobile Friendly**: Responsive design for all devices
- **🔗 Cross-References**: Automatic linking between related pages
- **📊 Analytics**: Usage tracking and content insights
- **🌐 SEO Optimized**: Meta tags and structured data
- **🛠️ Developer Friendly**: Straightforward to update and maintain

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- pnpm 9.3.0+

### Development Setup

1. **Install dependencies**:

   ```bash
   cd apps/docs
   pnpm install
   ```

2. **Start development server**:

   ```bash
   pnpm dev
   ```

3. **Access the documentation**:
   - http://localhost:3002

## 🏗️ Architecture

### Tech Stack

- **Framework**: Next.js 15 with App Router
- **Documentation**: Fumadocs for content management
- **Content**: MDX for rich documentation
- **Search**: Built-in full-text search
- **Styling**: Tailwind CSS with Fumadocs UI
- **Deployment**: Internally managed hosting

### Content Structure

```
apps/docs/
├── content/               # Documentation content
│   ├── docs/             # Main documentation pages
│   │   ├── getting-started/
│   │   ├── user-guide/
│   │   ├── api/
│   │   └── developer/
│   └── blog/             # Blog posts and announcements
├── components/           # Custom documentation components
├── lib/                  # Documentation utilities
│   ├── source.ts        # Content source configuration
│   └── mdx-components.tsx # Custom MDX components
├── app/                  # Next.js app structure
│   ├── (home)/          # Landing and overview pages
│   ├── docs/            # Documentation pages
│   └── api/             # Search API
└── source.config.ts     # Fumadocs configuration
```

## 📝 Writing Documentation

### Content Guidelines

- **Clear Structure**: Use headings, lists, and code blocks appropriately
- **Code Examples**: Include practical, runnable examples
- **Cross-References**: Link to related documentation
- **Up-to-Date**: Keep examples and instructions current
- **Accessible**: Write in clear, inclusive language

### MDX Features

```mdx
---
title: Getting Started
description: Learn how to get started with FortyOne
---

# Getting Started

Welcome to FortyOne! This guide explains the first workflow to configure.

## Before you begin

<Callout type="info">
  Ask a workspace administrator for an invitation before signing in.
</Callout>

Continue to the [workspace guide](/docs/product-guide/workspaces).
```

### Frontmatter

Use frontmatter to add metadata to your pages:

```yaml
---
title: Page Title
description: Brief description for SEO
icon: IconName
---
# Page Content
```

## 📦 Available Scripts

```bash
# Development
pnpm dev          # Start development server
pnpm build        # Build for production
pnpm start        # Start production server

# Content
pnpm postinstall  # Generate MDX content (runs automatically)
```

## 🚀 Deployment

The documentation site is deployed through the internally managed hosting
project. Production configuration and the `docs.fortyone.app` domain are owned
by the deployment platform; do not create independent public deployments.

## Updating documentation

### Adding New Pages

1. **Create MDX file** in appropriate directory under `content/`
2. **Add frontmatter** with title, description, and metadata
3. **Write content** following style guidelines
4. **Test locally** to ensure proper rendering
5. **Request internal review** before publishing

### Content Organization

```
content/docs/
├── getting-started/     # Onboarding guides
├── user-guide/         # Feature documentation
├── api/                # API reference
├── developer/          # Developer resources
└── troubleshooting/    # Common issues and solutions
```

### Style Guidelines

- **Headers**: Use sentence case for headings
- **Code**: Use backticks for inline code, triple backticks for blocks
- **Links**: Use relative links for internal documentation
- **Images**: Store in `public/` directory
- **Lists**: Use consistent formatting and indentation

## 🔧 Configuration

### Fumadocs Config

The `source.config.ts` file controls documentation behavior:

```typescript
import { createDocs } from "fumadocs-core";

export const source = createDocs({
  docs: {
    // Documentation pages
    dir: "content/docs",
    // URL prefix
    url: "/docs",
  },
  meta: {
    // Global metadata
    title: "FortyOne Docs",
    description: "Comprehensive documentation for FortyOne",
  },
});
```

### Custom Components

Add custom MDX components in `lib/mdx-components.tsx`:

```tsx
import { Callout } from "fumadocs-ui/components";

export const components = {
  Callout,
  // Add your custom components here
};
```

## 📊 Analytics & Insights

The documentation site includes analytics to understand:

- Most visited pages
- Search query patterns
- User engagement metrics
- Content effectiveness

Access analytics through your configured provider.

## Internal ownership

Documentation defects and publishing incidents are tracked through the private
engineering workflow. Customer-facing support remains available through the
channels published on `fortyone.app`.
