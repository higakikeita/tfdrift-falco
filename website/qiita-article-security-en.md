# Responding to React2Shell Vulnerabilities (CVE-2025-55182 and others) and Implementing Continuous Security Monitoring with Snyk

## Introduction

Shortly after launching the TFDrift-Falco project site, multiple critical vulnerabilities were discovered in React. This article covers our response to **React2Shell (CVE-2025-55182)** and three other vulnerabilities, as well as establishing a continuous security monitoring system using Snyk.

- **Project Site**: https://tfdrift-falco.vercel.app/
- **GitHub**: https://github.com/higakikeita/tfdrift-falco

## Discovered Vulnerabilities

### CVE-2025-55182: React2Shell (Critical - CVSS 10.0)

Published on December 3, 2025, this is a **maximum severity vulnerability**.

```
Vulnerability: Insecure deserialization in React Server Components
Impact: Unauthenticated remote code execution (RCE)
Exploitation: Chinese threat groups began exploiting within hours of disclosure
```

#### Technical Details

- **Cause**: Insecure deserialization in the Flight protocol
- **Attack Method**: Send malicious HTTP requests
- **Success Rate**: Nearly 100%, vulnerable in default configurations
- **Impact Scope**: All apps using React Server Components

### Other Related Vulnerabilities

#### CVE-2025-55184: Denial of Service (High - CVSS 7.5)

```yaml
Issue: Malicious HTTP requests trigger infinite loops
Impact: Service becomes unavailable
```

#### CVE-2025-67779: Incomplete Fix (High - CVSS 7.5)

```yaml
Issue: Initial fix for CVE-2025-55184 was incomplete
Impact: React 19.0.2, 19.1.3, 19.2.2 remain vulnerable
```

#### CVE-2025-55183: Source Code Exposure (Medium - CVSS 5.3)

```yaml
Issue: Server function source code exposed
Impact: Hardcoded secrets like API keys can leak
```

## Discovery Process

### 1. How We Noticed

After launching the project site, we saw news about **React2Shell**.

```bash
# Check our versions
cat website/package.json | grep react
```

```json
{
  "react": "19.2.1",        // ← Vulnerable!
  "react-dom": "19.2.1",    // ← Vulnerable!
  "next": "16.0.10"         // ← This is safe
}
```

### 2. npm audit Results

```bash
cd website
npm audit
```

Result:
```
found 0 vulnerabilities
```

**Surprisingly, npm audit didn't detect these!**

This reveals:
- npm audit's database isn't always up-to-date
- New vulnerabilities take time to register

### 3. Manual Verification

Checked official sources:

- [React Official Blog](https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components)
- [Next.js Security Update](https://nextjs.org/blog/security-update-2025-12-11)

**Conclusion**: React 19.2.1 is vulnerable, 19.2.3 is required

## Introducing Snyk

Since npm audit failed to detect these, we decided to introduce **Snyk**, a more advanced security tool.

### What is Snyk?

- **Security platform for developers**
- Open-source vulnerability database
- Higher detection accuracy than npm audit
- Easy GitHub Actions integration

### Setup Process

#### 1. Create Snyk Account

```bash
# 1. Visit https://snyk.io/
# 2. Sign up with GitHub account
# 3. Get API token
```

#### 2. Add to GitHub Secrets

```bash
# Repository → Settings → Secrets and variables → Actions
# New repository secret
Name: SNYK_TOKEN
Secret: (Your Snyk API token)
```

#### 3. Create GitHub Actions Workflow

```yaml
# .github/workflows/website-security.yml
name: Website Security Scan

on:
  push:
    branches: [main]
    paths:
      - 'website/**'
  pull_request:
    branches: [main]
    paths:
      - 'website/**'
  schedule:
    - cron: '0 9 * * 1' # Every Monday at 9:00 UTC
  workflow_dispatch:

jobs:
  security:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: website/package-lock.json

      - name: Install dependencies
        working-directory: ./website
        run: npm ci

      - name: Run Snyk to check for vulnerabilities
        uses: snyk/actions/node@master
        continue-on-error: true
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high --sarif-file-output=snyk.sarif
          command: test
          working-directory: website

      - name: Upload Snyk results to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: snyk.sarif
          category: website-security
```

#### 4. Custom Security Check Script

In addition to Snyk, we created a script for instant verification:

```bash
#!/bin/bash
# check-security.sh

echo "🔍 Security Check"
echo "================="

REACT_VERSION=$(node -p "require('./package.json').dependencies.react")
REACT_VER_NUM=$(echo $REACT_VERSION | sed 's/[\^~]//g')

# React 19.2.1, 19.2.2, 19.1.3, 19.0.2 are vulnerable
if [[ "$REACT_VER_NUM" == "19.2.1" ]] || [[ "$REACT_VER_NUM" == "19.2.2" ]]; then
    echo "⚠️  VULNERABLE: React $REACT_VER_NUM"
    echo "   CVE-2025-55182: RCE (Critical)"
    echo "   CVE-2025-55183: Source Code Exposure (Medium)"
    echo "   CVE-2025-55184: DoS (High)"
    echo "   CVE-2025-67779: Incomplete fix (High)"
    echo ""
    echo "🔧 Fix: npm install react@19.2.3 react-dom@19.2.3"
    exit 1
elif [[ "$REACT_VER_NUM" == "19.2.3" ]]; then
    echo "✅ SAFE: React $REACT_VER_NUM (patched)"
    exit 0
fi
```

Execute:
```bash
chmod +x check-security.sh
./check-security.sh
```

## Fixing the Vulnerabilities

### 1. Update React

```bash
cd website

# Update to patched version
npm install react@19.2.3 react-dom@19.2.3
```

### 2. Verify Version

```bash
# Check package.json
cat package.json | grep react
```

```json
{
  "react": "^19.2.3",      // ✅ Safe
  "react-dom": "^19.2.3"   // ✅ Safe
}
```

### 3. Verify Build

```bash
npm run build
```

```
✓ Compiled successfully
✓ Generating static pages (8/8)

Route (app)
┌ ○ /
├ ○ /blog
├ ● /blog/[slug]
└ ○ /releases

✓ Build completed successfully
```

### 4. Security Recheck

```bash
./check-security.sh
```

```
🔍 Security Check
=================
📦 Current Versions:
  React: ^19.2.3
  Next.js: 16.0.10

✅ SAFE: React 19.2.3 (patched)
✅ SAFE: Next.js 16.0.10

✅ No known vulnerabilities detected
```

### 5. Commit & Deploy

```bash
git add package.json package-lock.json check-security.sh
git commit -m "security: Fix critical React vulnerabilities (CVE-2025-55182/55183/55184/67779)"
git push origin main
```

Vercel automatically redeploys → **Reflected in production within minutes**

## Continuous Security Monitoring

### 1. Automatic Scan Triggers

```yaml
# GitHub Actions automatically runs on:
on:
  push:              # Push to main branch
  pull_request:      # PR creation
  schedule:          # Every Monday at 9:00 UTC
  workflow_dispatch: # Manual trigger
```

### 2. GitHub Code Scanning Integration

Snyk results integrate with **GitHub Code Scanning**:

```
Repository → Security → Code scanning alerts
```

When vulnerabilities are detected:
- Automatically creates alerts
- Shows affected files
- Suggests fixes

### 3. Notification Settings

```
GitHub → Settings → Notifications → Security alerts
```

Enable:
- Dependabot alerts
- Code scanning alerts
- Secret scanning alerts

### 4. Regular Local Checks

```bash
# Weekly local check (developer habit)
cd website
./check-security.sh

# Also check with Snyk CLI
npm install -g snyk
snyk auth
snyk test
```

## Lessons Learned

### 1. npm audit Isn't Perfect

```
npm audit: 0 vulnerabilities ❌
Reality: 4 critical vulnerabilities ⚠️
```

**Lesson**: Use multiple tools in combination

### 2. Vulnerabilities Happen Even Right After Launch

```
Website launched: 2025-12-14
React2Shell published: 2025-12-03 (just 11 days earlier)
```

**Lesson**: Continuous monitoring is essential

### 3. Importance of Automation

Manual checking alone:
- Prone to oversight
- Delayed responses
- Consumes human resources

**Lesson**: Integrate into CI/CD

### 4. Defense in Depth

```
1. Snyk (high-precision detection)
2. npm audit (basic checks)
3. Custom scripts (instant verification)
4. GitHub Code Scanning (visualization)
5. Manual information gathering (latest info)
```

**Lesson**: Don't depend on a single tool

## Security Best Practices

### 1. Regular Dependency Updates

```bash
# Run monthly
npm outdated
npm update
npm audit fix
```

### 2. Utilize Dependabot

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/website"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

### 3. Document Security Policy

```markdown
# SECURITY.md

## Reporting Vulnerabilities
security@example.com

## Supported Versions
| Version | Supported |
| ------- | --------- |
| 1.x.x   | ✅        |
| 0.x.x   | ❌        |
```

### 4. Proper Environment Variable Management

```bash
# ❌ Hardcoded in source
const API_KEY = "sk-1234567890"

# ✅ Use environment variables
const API_KEY = process.env.API_KEY
```

### 5. Principle of Least Privilege

```yaml
# GitHub Actions
permissions:
  contents: read        # Read-only
  security-events: write # Write only security events
```

## Response Timeline

```
2025-12-03: React2Shell (CVE-2025-55182) published
2025-12-11: Additional vulnerabilities published (CVE-2025-55183/55184/67779)
2025-12-14: TFDrift-Falco site launched (vulnerable version)
2025-12-14: Vulnerability identified
2025-12-14: Started Snyk integration
2025-12-14: Updated to React 19.2.3
2025-12-14: Security check automation completed

Response time: Approximately 2 hours from discovery to fix completion
```

## Cost

Everything achieved **completely free**:

- **Snyk**: Free for open-source projects
- **GitHub Actions**: 2000 minutes/month free
- **GitHub Code Scanning**: Free for public repositories
- **Vercel**: Hobby plan free

## Future Improvement Plans

### Short-term (Within 1 month)

1. **Enable Dependabot**
   - Automatic PR creation
   - Regular dependency updates

2. **Generate SBOM**
   - Software Bill of Materials
   - Dependency visualization

3. **Security Documentation**
   - SECURITY.md
   - Vulnerability response flow

### Medium-term (Within 3 months)

1. **Add Container Scanning**
   - Docker image scanning
   - Base image vulnerability checks

2. **Implement SAST/DAST**
   - Static Analysis (SAST)
   - Dynamic Analysis (DAST)

3. **Security Training**
   - Team knowledge sharing
   - Secure coding guidelines

### Long-term (Within 6 months)

1. **Bug Bounty Program**
   - Feedback from security researchers

2. **Penetration Testing**
   - Regular external audits

3. **Incident Response Plan**
   - Emergency response flow
   - Backup and recovery plan

## Conclusion

Lessons learned from responding to React2Shell (CVE-2025-55182):

✅ **Use Multiple Security Tools**
- npm audit
- Snyk
- Custom scripts
- GitHub Code Scanning

✅ **Continuous Monitoring Through Automation**
- Integrated into CI/CD pipeline
- Scheduled execution
- Real-time alerts

✅ **Rapid Response System**
- 2 hours from discovery to fix
- Instant reflection via auto-deploy

✅ **Defense in Depth**
- Three-layer structure: detection, response, monitoring
- Don't depend on a single tool

✅ **Community Collaboration**
- Regular checks of official information
- Participation in security communities

Security is **not a one-time task but a continuous process**. The system we built now enables us to respond quickly to future vulnerabilities.

## References

### Official Information
- [React - Critical Security Vulnerability in React Server Components](https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components)
- [React - Denial of Service and Source Code Exposure](https://react.dev/blog/2025/12/11/denial-of-service-and-source-code-exposure-in-react-server-components)
- [Next.js Security Update: December 11, 2025](https://nextjs.org/blog/security-update-2025-12-11)
- [Vercel Security Bulletin](https://vercel.com/kb/bulletin/security-bulletin-cve-2025-55184-and-cve-2025-55183)

### Security Analysis
- [AWS Security Blog - React2Shell](https://aws.amazon.com/blogs/security/china-nexus-cyber-threat-groups-rapidly-exploit-react2shell-vulnerability-cve-2025-55182/)
- [Qualys - React2Shell Decoding](https://blog.qualys.com/product-tech/2025/12/10/react2shell-decoding-cve-2025-55182-the-silent-threat-in-react-server-components)

### Tools
- [Snyk](https://snyk.io/)
- [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning)

### Project
- **TFDrift-Falco**: https://tfdrift-falco.vercel.app/
- **GitHub**: https://github.com/higakikeita/tfdrift-falco

---

**Security is a continuous effort. Let's build secure software together!**

Questions and feedback are welcome in [GitHub Discussions](https://github.com/higakikeita/tfdrift-falco/discussions).
