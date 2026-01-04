# Open Source Readiness Plan

## Overview

This document outlines all the changes and additions needed to make the FortyOne project open-source ready.

---

## ✅ CRITICAL - Required Before Publishing

### 1. Add LICENSE File

**Status:** ✅ Complete - MIT License + Commons Clause
**Action:** Created `LICENSE` file in root directory

**Chosen License:** MIT License with Commons Clause addendum

**Why this license:**

- Allows free personal and non-commercial use
- Prevents commercial exploitation (SaaS offerings, paid hosting, etc.)
- Requires commercial entities to obtain separate licensing
- Maintains open-source development benefits

**License Details:**

- **Base License:** MIT (permissive, widely compatible)
- **Commercial Restriction:** Commons Clause (prohibits selling/offering as service)
- **Licensor:** Complexus LLC

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

**Status:** ✅ Complete
**Action:** Created environment variable examples

Created `.env.example` and `.env` files for apps:

- ✅ `apps/landing/.env.example` & `.env` - Domain, API, analytics, auth, AI services, OAuth, Sentry
- ✅ `apps/projects/.env.example` & `.env` - Domain, API, analytics, auth, AI services, Sentry, Azure storage
- ✅ `apps/mobile/.env.example` & `.env` - API URL for Expo
- ℹ️ `apps/docs/` - No environment variables needed (static docs site)

**⚠️ IMPORTANT:** Update `NEXT_PUBLIC_SENTRY_DSN=your_actual_sentry_dsn_here` in `.env` files with your real Sentry DSN before deploying.

---

### 4. Move Hardcoded Sentry DSN to Environment Variable

**Status:** ✅ Complete - Security Issue Resolved
**Action:** Updated Sentry configuration files to use environment variables

**Files updated:**

- ✅ `apps/projects/sentry.server.config.ts` - DSN now uses `process.env.NEXT_PUBLIC_SENTRY_DSN`
- ✅ `apps/projects/sentry.edge.config.ts` - DSN now uses `process.env.NEXT_PUBLIC_SENTRY_DSN`

**Security Fix Applied:**

```typescript
// Before (EXPOSED SECRET):
dsn: "https://1731711ff88c9c60022dadf2d2e85381@o4508848135077888.ingest.us.sentry.io/4508848139075584",

// After (SECURE):
dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
```

**Impact:** Removed sensitive API credentials from codebase. Developers must now configure `NEXT_PUBLIC_SENTRY_DSN` in their environment.

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

1. ✅ Add LICENSE (MIT + Commons Clause)
2. ✅ Update package.json `private` flags (root, apps, lib)
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
