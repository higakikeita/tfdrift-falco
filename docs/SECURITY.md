# Security Policy

## Supported Versions

We take security seriously and will address vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in TFDrift-Falco, please report it responsibly by following one of these methods:

### 1. GitHub Security Advisories (Preferred)

1. Go to the [Security tab](https://github.com/higakikeita/tfdrift-falco/security)
2. Click "Report a vulnerability"
3. Fill out the form with details about the vulnerability

### 2. Direct Contact

Send an email to:
- **X (Twitter) DM**: [@keitah0322](https://x.com/keitah0322)
- **GitHub**: Create a draft security advisory in this repository

### What to Include

Please include the following information in your report:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** of the vulnerability
- **Suggested fix** (if you have one)
- **Your contact information** for follow-up questions

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: 1-7 days
  - High: 7-14 days
  - Medium: 14-30 days
  - Low: 30-90 days

## Security Considerations for TFDrift-Falco

Since TFDrift-Falco handles sensitive infrastructure data, please be aware of these security considerations:

### 1. Credentials Management

- **Never commit AWS credentials** to the repository
- Use IAM roles and instance profiles when possible
- Rotate credentials regularly
- Follow the principle of least privilege

### 2. Terraform State Files

- Terraform state files may contain **sensitive data**
- Always use encrypted remote backends (S3 with KMS, Terraform Cloud)
- Never commit `.tfstate` files to version control
- Restrict access to state files using IAM policies

### 3. CloudTrail Data

- CloudTrail logs may contain **PII and sensitive API calls**
- Ensure proper encryption of CloudTrail S3 buckets
- Implement proper access controls on SQS queues
- Consider data retention policies

### 4. Network Security

- Run TFDrift-Falco in a **private subnet** when possible
- Use VPC endpoints for AWS service access
- Implement proper security group rules
- Enable VPC Flow Logs for network monitoring

### 5. Notification Channels

- **Webhook URLs** (Slack, Discord) should be treated as secrets
- Use environment variables or secrets management tools
- Never hardcode webhook URLs in configuration files
- Rotate webhook URLs if compromised

## Security Best Practices

When deploying TFDrift-Falco:

1. **Run with minimal IAM permissions**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:GetObject",
           "sqs:ReceiveMessage",
           "sqs:DeleteMessage"
         ],
         "Resource": [
           "arn:aws:s3:::your-cloudtrail-bucket/*",
           "arn:aws:sqs:*:*:cloudtrail-events"
         ]
       }
     ]
   }
   ```

2. **Enable encryption at rest and in transit**
   - Use TLS for all API calls
   - Encrypt Terraform state files
   - Enable S3 bucket encryption

3. **Audit and logging**
   - Enable CloudTrail for all AWS accounts
   - Monitor TFDrift-Falco's own actions
   - Set up alerts for suspicious activity

4. **Regular updates**
   - Keep TFDrift-Falco up to date
   - Monitor security advisories
   - Update dependencies regularly

## Known Security Limitations

### Current Version (0.1.x)

- State file encryption is not implemented (use encrypted remote backends)
- No built-in secrets scanning for Terraform files
- CloudTrail events are processed without additional verification
- No rate limiting on notification channels

These limitations are tracked in our [Security Roadmap](https://github.com/higakikeita/tfdrift-falco/issues).

## Security Roadmap

Planned security enhancements:

- [ ] Secrets scanning for Terraform state files
- [ ] End-to-end encryption for notification payloads
- [ ] Signature verification for CloudTrail events
- [ ] Rate limiting and throttling
- [ ] Audit logging for TFDrift-Falco actions
- [ ] RBAC support for multi-tenant deployments

## Acknowledgments

We appreciate the security research community and will publicly acknowledge researchers who responsibly disclose vulnerabilities (with their permission).

### Hall of Fame

<!-- Security researchers will be listed here -->

*No vulnerabilities reported yet.*

## Further Reading

- [AWS Security Best Practices](https://aws.amazon.com/security/best-practices/)
- [Terraform Security Best Practices](https://www.terraform.io/docs/cloud/guides/recommended-practices/index.html)
- [Falco Security](https://falco.org/docs/security/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

**Thank you for helping keep TFDrift-Falco and its users safe!**
