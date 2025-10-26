# Open Source Readiness Plan

## Overview

This document outlines all the changes and additions needed to make the FortyOne project open-source ready.

---

## ✅ CRITICAL - Required Before Publishing

### 1. Add LICENSE File

**Status:** ❌ Missing  
**Action:** Create a `LICENSE` file in the root directory

Choose an appropriate open-source license:

- **MIT** (Recommended) - Most permissive, widely used
- **Apache 2.0** - More corporate-friendly, includes patent protection
- **GPL** - Copyleft, requires derivative works to be open source

Example MIT License structure:

```
MIT License

Copyright (c) 2025 FortyOne

Permission is hereby granted, free of charge, to any person obtaining a copy...
```

---

### 2. Remove "private" flag from package.json

**Status:** ❌ Currently set to `true`  
**Action:** Change `"private": true` to `"private": false` in:

- [ ] `package.json` (root)
- [ ] `apps/landing/package.json`
- [ ] `apps/projects/package.json`
- [ ] `apps/mobile/package.json`
- [ ] `apps/docs/package.json`

**Why:** npm/pnpm will refuse to publish private packages publicly.

---

### 3. Create .env.example Files

**Status:** ❌ Missing  
**Action:** Create environment variable examples

#### Root `.env.example`

```bash
# Application URLs
NEXT_PUBLIC_API_URL=https://api.example.com
NEXT_PUBLIC_DOMAIN=fortyone.app
EXPO_PUBLIC_API_URL=https://api.example.com

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

# Error Tracking (Optional)
NEXT_PUBLIC_SENTRY_DSN=your_sentry_dsn

# Google OAuth (for landing app)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

Also create `.env.example` files in:

- [ ] `apps/landing/` (if app-specific vars needed)
- [ ] `apps/projects/` (if app-specific vars needed)
- [ ] `apps/mobile/` (if app-specific vars needed)

---

### 4. Move Hardcoded Sentry DSN to Environment Variable

**Status:** ⚠️ Critical Security Issue  
**Action:** Update these files:

**Files to update:**

- [ ] `apps/projects/sentry.server.config.ts` (line 8)
- [ ] `apps/projects/sentry.edge.config.ts` (line 8)

**Current code:**

```typescript
dsn: "https://1731711ff88c9c60022dadf2d2e85381@o4508848135077888.ingest.us.sentry.io/4508848139075584",
```

**Should be:**

```typescript
dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
```

**Why:** This is a sensitive identifier that should not be committed to the repository.

---

### 5. Update Email Addresses

**Status:** ❌ Contains personal/organizational emails  
**Action:** Replace all instances of:

**Files containing email references:**

- [ ] `apps/landing/src/modules/contact/support.tsx`
- [ ] `apps/landing/src/content/terms.md`
- [ ] `apps/landing/src/content/privacy-policy.md`
- [ ] `apps/docs/content/docs/help-and-support/contact-us.mdx`
- [ ] `apps/landing/src/app/(marketing)/contact/json-ld.tsx`
- [ ] `apps/projects/src/components/shared/sidebar/sidebar.tsx`
- [ ] `apps/projects/src/modules/settings/workspace/billing/index.tsx`
- [ ] `apps/projects/src/modules/settings/workspace/billing/components/plans.tsx`
- [ ] `apps/landing/src/components/shared/json-ld.tsx`

**Replace:**

- `support@complexus.app` → Your support email or consider removing for open-source
- `sales@complexus.app` → Your sales email or consider removing for open-source
- Websites or business emails mentioned in TODOs

**Options:**

1. Use a generic email like `support@fortyone.app` (if you control this domain)
2. Use GitHub discussions/email for support
3. Remove contact methods altogether (limit to GitHub issues)

---

## 📝 DOCUMENTATION - Highly Recommended

### 6. Create CONTRIBUTING.md

**Status:** ❌ Missing  
**Action:** Create a `CONTRIBUTING.md` file

**Suggested content:**

- Code of conduct or link to one
- How to report bugs (use GitHub issues)
- How to suggest features
- Development environment setup
- Git workflow (branching strategy)
- Code style guidelines
- Testing requirements
- How to submit a pull request
- Response time expectations

**Template resources:**

- [Contributor Covenant](https://www.contributor-covenant.org/)
- [GitHub's Contributing Guide Template](https://github.com/github/docs/blob/main/CONTRIBUTING.md)

---

### 7. Create SECURITY.md

**Status:** ❌ Missing  
**Action:** Create a `SECURITY.md` file

**Suggested content:**

- Security policy
- How to report vulnerabilities
- What to include in a security report
- Supported versions
- Response timeline
- Security updates policy

**Template:**

```markdown
# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version  | Supported          |
| -------- | ------------------ |
| Latest   | :white_check_mark: |
| < Latest | :x:                |

## Reporting a Vulnerability

Please report security vulnerabilities via [email/security advisory on GitHub].
We will acknowledge your email, and send a more detailed daily response within 48 hours.
```

---

### 8. Update README Files

**Status:** ⚠️ Needs improvement  
**Action:** Enhance existing READMEs

#### Root `README.md`

**Current:** Good setup guide  
**Add:**

- [ ] Clear project description at the top
- [ ] Features list
- [ ] Demo/screenshot section
- [ ] Badges (build status, license, version)
- [ ] Link to documentation
- [ ] Contributing section
- [ ] Code of conduct
- [ ] License section
- [ ] Contact/support information
- [ ] Architecture overview

#### App-specific READMEs

**Files:**

- [ ] `apps/docs/README.md` - Very generic
- [ ] `apps/landing/README.md` - Very generic
- [ ] `apps/projects/README.md` - Very generic

**Add:** App-specific information, setup, and deployment instructions

---

### 9. Document External Dependencies

**Status:** ⚠️ Partially documented  
**Action:** Create comprehensive service documentation

**External services used:**

- [ ] **PostHog** - Analytics
- [ ] **Sentry** - Error tracking
- [ ] **Google Analytics** - Analytics
- [ ] **Google OAuth** - Authentication
- [ ] **OpenAI** - AI features
- [ ] **Google Gemini** - AI features
- [ ] **Cal.com** - Scheduling
- [ ] **Azure Blob Storage** - File storage
- [ ] **Backend API** - Custom backend (not in this repo)

**Documentation needed:**

- Which services are required vs. optional
- How to obtain API keys
- Free tier availability
- Alternative services for open-source users
- Costs/warnings about usage

---

## 🎨 BRANDING & CONSISTENCY

### 10. Resolve Branding Inconsistencies

**Status:** ⚠️ Mixed branding  
**Issue:** Project references coaches both "complexus" and "fortyone"

**Files to check:**

- [ ] Repository structure has `apps/complexus/` directories
- [ ] README references "fortyone.tech"
- [ ] Some files reference "complexus"
- [ ] Domain: fortyone.lc vs complexus.app

**Action:** Decide on single branding and update:

1. Directory names
2. References in code
3. Documentation
4. Domain mentions
5. Package names if needed

---

## 🚀 OPTIONAL - Nice to Have

### 11. GitHub Repository Configuration

**Status:** Recommended  
**Action:** Add GitHub-specific files

- [ ] **.github/ISSUE_TEMPLATE/** - Issue templates
  - Bug report template
  - Feature request template
  - Question template
- [ ] **.github/PULL_REQUEST_TEMPLATE.md** - PR template
- [ ] **.github/CODEOWNERS** - Automatically request reviews
- [ ] **.github/workflows/** - CI/CD setup
- [ ] **.github/dependabot.yml** - Dependencies updates
- [ ] **FUNDING.yml** - Funding/sponsorship information

---

### 12. Code Quality

**Status:** Good  
**Optional improvements:**

- [ ] Add more comprehensive tests
- [ ] Setup CI/CD pipeline
- [ ] Code coverage reporting
- [ ] Automated code quality checks

---

### 13. Additional Documentation

**Status:** Optional  
**Consider:**

- [ ] Architecture decision records (ADRs)
- [ ] API documentation
- [ ] Deployment guides
- [ ] Troubleshooting guide
- [ ] FAQ section

---

## 🔍 FINAL CHECKLIST

Before making the repository public, verify:

### Security

- [ ] No API keys or secrets in code
- [ ] No database connection strings
- [ ] No personal information
- [ ] No proprietary code (unless intended)
- [ ] Sensitive credentials are in `.env.example` or documentation only

### Legal

- [ ] LICENSE file added
- [ ] Copyright notices correct
- [ ] Third-party licenses documented (if any)
- [ ] Terms of service updated (if applicable)
- [ ] Privacy policy updated (if applicable)

### Code

- [ ] `private: false` in all package.json files
- [ ] `.env.example` files present
- [ ] No hardcoded secrets
- [ ] Code is properly formatted
- [ ] Tests are passing

### Documentation

- [ ] README is complete
- [ ] CONTRIBUTING.md exists
- [ ] SECURITY.md exists
- [ ] Environment variables documented
- [ ] External services documented
- [ ] Setup instructions are clear

### Branding

- [ ] Consistent branding throughout
- [ ] Contact information updated
- [ ] No personal information exposed

---

## 📊 Summary of Required Changes

**Critical (Must Fix):**

1. ✅ Add LICENSE
2. ✅ Update package.json `private` flags
3. ✅ Create .env.example files
4. ✅ Move hardcoded Sentry DSN to env vars
5. ✅ Update email addresses

**Important (Should Have):** 6. ✅ Create CONTRIBUTING.md 7. ✅ Create SECURITY.md 8. ✅ Enhance README.md files 9. ✅ Document external services 10. ✅ Resolve branding inconsistencies

**Optional (Nice to Have):** 11. ✅ GitHub templates and workflows 12. ✅ Additional tests and code quality 13. ✅ Extra documentation

---

## 🎯 Next Steps

1. Review this document
2. Prioritize the critical items
3. Work through each section
4. Test the setup with a fresh clone
5. Do a final security review
6. Make the repository public
7. Announce the open-source release

---

## 📞 Questions or Need Help?

If you have questions about any of these steps or need clarification, feel free to ask!

Good luck with your open-source journey! 🚀
