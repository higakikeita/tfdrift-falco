# Security Policy

## Supported Versions

Currently, security updates are provided for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :x:                |

## Security Scanning

This project uses multiple security scanning tools:

### 1. Snyk
- **Purpose**: Vulnerability scanning for dependencies
- **Runs**: On every push, pull request, and weekly
- **Configuration**: Requires `SNYK_TOKEN` secret

To set up Snyk for your fork:
1. Sign up at [snyk.io](https://snyk.io/)
2. Get your API token from Account Settings
3. Add it as `SNYK_TOKEN` in your repository secrets (Settings → Secrets and variables → Actions)

### 2. GoSec
- **Purpose**: Security audit of Go source code
- **Runs**: On every push and pull request
- **Configuration**: No secrets required

### 3. Nancy
- **Purpose**: OSS dependency vulnerability scanner
- **Runs**: On every push and pull request
- **Configuration**: No secrets required

## Reporting a Vulnerability

If you discover a security vulnerability, please follow these steps:

1. **DO NOT** open a public issue
2. Email the maintainers at [security email - to be added]
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We aim to respond to security reports within 48 hours.

## Security Best Practices

When contributing to this project:

1. **Dependencies**: Keep dependencies up to date
2. **Secrets**: Never commit secrets, tokens, or credentials
3. **Input Validation**: Validate all external input
4. **Error Handling**: Avoid exposing sensitive information in errors
5. **Permissions**: Use minimum required file permissions (0600 for sensitive files)

## Security Features

- Dry-run mode by default
- Input validation for all configuration
- Secure file permissions for state files
- No exposure of AWS credentials in logs
- SARIF output for security scan results
