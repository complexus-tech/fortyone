# Security Policy

## Supported Versions

We take security seriously and actively maintain security updates. The following versions receive security updates:

| Version | Supported          | Security Updates |
| ------- | ------------------ | ---------------- |
| Latest  | :white_check_mark: | :white_check_mark: |
| < Latest| :x:                | :white_check_mark: |

**Security updates are provided for all released versions**, but new features are only added to the latest version.

## Reporting a Vulnerability

If you discover a security vulnerability in FortyOne, please help us by reporting it responsibly.

### How to Report

**DO NOT** create public GitHub issues for security vulnerabilities.

Instead, please report security issues by emailing: **hello@complexus.tech**

### What to Include

Please include the following information in your report:

- **Description**: A clear description of the vulnerability
- **Steps to reproduce**: Detailed steps to reproduce the issue
- **Impact**: Potential impact and severity of the vulnerability
- **Affected versions**: Which versions are affected
- **Environment**: Any specific environment details
- **Contact information**: How we can reach you for follow-up

### Response Timeline

We will acknowledge your report within **48 hours** and provide a more detailed response within **7 days** indicating our next steps.

We will keep you informed about our progress throughout the process of fixing the vulnerability.

### Disclosure Policy

- We will credit you (if desired) once the vulnerability is fixed
- We will not disclose vulnerability details until a fix is available
- We follow responsible disclosure practices

## Security Best Practices

### For Contributors

When contributing code, please:

- Avoid hardcoding sensitive information
- Use environment variables for secrets
- Follow secure coding practices
- Add input validation and sanitization
- Use parameterized queries for database operations

### For Users

To keep your FortyOne installation secure:

- Keep dependencies updated
- Use strong, unique passwords
- Enable two-factor authentication
- Regularly backup your data
- Monitor for security updates

## Security Updates

### How We Handle Updates

1. **Assessment**: Security team assesses the vulnerability
2. **Fix Development**: Develop and test security fixes
3. **Release**: Release security updates to all supported versions
4. **Communication**: Notify users through security advisories
5. **Public Disclosure**: Make vulnerability details public after fixes are available

### Security Advisories

Security advisories will be published on:
- GitHub Security Advisories
- Our security mailing list
- Project documentation

## Contact

For security-related questions or concerns:
- **Email**: security@fortyone.app
- **PGP Key**: Available upon request for encrypted communications

## Recognition

We appreciate security researchers who help keep FortyOne safe. With your permission, we'll acknowledge your contribution in our security hall of fame.

## Disclaimer

This security policy applies to the FortyOne software and related components. Third-party integrations and services may have their own security policies.