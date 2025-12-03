# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < main  | :x:                |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of these methods:

1. **GitHub Security Advisories** (Preferred)
   - Go to the [Security tab](https://github.com/DrBlury/Event-Driven-Service-Example/security/advisories)
   - Click "Report a vulnerability"
   - Fill out the form with details

2. **Email**
   - Send details to the repository maintainer
   - Include "SECURITY" in the subject line

### What to Include

Please include the following in your report:

- **Description**: A clear description of the vulnerability
- **Impact**: What an attacker could achieve
- **Steps to Reproduce**: Detailed steps to reproduce the issue
- **Affected Versions**: Which versions are affected
- **Suggested Fix**: If you have a suggested fix, please include it

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution Target**: Within 30 days for critical issues

### What to Expect

1. **Acknowledgment**: We'll acknowledge receipt of your report
2. **Investigation**: We'll investigate and determine the severity
3. **Fix Development**: We'll develop a fix if the issue is confirmed
4. **Disclosure**: We'll coordinate disclosure timing with you
5. **Credit**: We'll credit you in the release notes (unless you prefer anonymity)

## Security Best Practices

When using this service, please follow these security best practices:

### Configuration

- Never commit secrets or credentials to the repository
- Use environment variables or secret management tools
- Rotate credentials regularly
- Use TLS/HTTPS in production

### Deployment

- Keep dependencies updated (Dependabot is enabled)
- Run security scans before deployment
- Use the principle of least privilege
- Enable audit logging

### Monitoring

- Monitor for unusual activity
- Set up alerts for security events
- Review logs regularly

## Security Features

This project includes the following security measures:

- **Dependency Scanning**: Dependabot for automated updates
- **Code Scanning**: gosec and govulncheck in CI
- **Secret Scanning**: Trufflehog for leaked secrets
- **Vulnerability Scanning**: Trivy and Grype scans
- **SBOM Generation**: Software Bill of Materials for supply chain security

## Acknowledgments

We appreciate the security research community's efforts in keeping our project secure. Contributors who responsibly disclose vulnerabilities will be acknowledged here:

<!-- Security researchers who have helped improve our security -->
