# Contributing to TFDrift-Falco

First off, thank you for considering contributing to TFDrift-Falco! It's people like you that make TFDrift-Falco such a great tool.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [security@example.com](mailto:security@example.com).

### Our Standards

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- Docker (for testing)
- AWS CLI (for AWS integration testing)
- gcloud CLI (for GCP integration testing - v0.5.0+)
- Basic knowledge of Terraform and Falco
- Basic knowledge of cloud platforms (AWS, GCP)

### Types of Contributions

We love contributions from community members, just like you! Here are ways you can contribute:

- 🐛 **Bug reports** - Found a bug? Let us know!
- ✨ **Feature requests** - Have an idea? We'd love to hear it!
- 📝 **Documentation** - Help improve our docs
- 🔧 **Code contributions** - Fix bugs or implement features
- 🧪 **Tests** - Help us improve test coverage
- 🌍 **Translations** - Help translate docs to other languages

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/tfdrift-falco.git
cd tfdrift-falco

# Add upstream remote
git remote add upstream https://github.com/higakikeita/tfdrift-falco.git
```

### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Build the Project

```bash
# Build binary
go build -o tfdrift ./cmd/tfdrift

# Or use Make
make build
```

### 4. Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### 5. Run Locally

```bash
# Copy example config
cp examples/config.yaml config.yaml

# Edit config.yaml with your settings
vim config.yaml

# Run in dry-run mode
./tfdrift --config config.yaml --dry-run
```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

**Bug Report Template:**

```markdown
**Description**
A clear and concise description of the bug.

**To Reproduce**
Steps to reproduce the behavior:
1. Configure '...'
2. Run command '...'
3. See error

**Expected Behavior**
What you expected to happen.

**Actual Behavior**
What actually happened.

**Environment**
- OS: [e.g. Ubuntu 22.04]
- Go Version: [e.g. 1.24]
- TFDrift-Falco Version: [e.g. 0.5.0]
- Cloud Provider: [e.g. AWS, GCP, or both]
- Terraform Version: [e.g. 1.6.0]

**Logs**
```
Paste relevant logs here
```

**Additional Context**
Add any other context about the problem here.
```

### Suggesting Features

Feature requests are welcome! Please provide:

1. **Use case** - Why do you need this feature?
2. **Proposed solution** - How should it work?
3. **Alternatives considered** - What other solutions did you consider?
4. **Additional context** - Any other relevant information

**Feature Request Template:**

```markdown
**Is your feature request related to a problem?**
A clear and concise description of what the problem is.

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions.

**Example Configuration**
```yaml
# How would this feature be configured?
```

**Additional context**
Add any other context or screenshots about the feature request.
```

## Coding Standards

### Go Style Guidelines

We follow the standard Go coding conventions:

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` to format code
- Use `golangci-lint` for linting
- Keep functions focused and small
- Write clear, descriptive variable names
- Add comments for exported functions and types

### Code Organization

```
pkg/
├── aws/             # AWS-specific parsers and logic
├── gcp/             # GCP-specific parsers and logic (v0.5.0+)
├── cloudtrail/      # CloudTrail event collection
├── config/          # Configuration management
├── detector/        # Core drift detection logic
├── falco/           # Falco integration
├── notifier/        # Notification handling
├── terraform/       # Terraform state management
│   └── backend/     # State backend implementations (S3, GCS, local)
└── types/           # Common types and interfaces
```

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to load state: %w", err)
}

// Bad: Return raw errors
if err != nil {
    return err
}
```

### Logging

```go
// Use structured logging
log.WithFields(log.Fields{
    "resource_id": resourceID,
    "event_type":  eventType,
}).Info("Processing event")

// Don't use fmt.Println()
```

## Testing Guidelines

### Unit Tests

- Write unit tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests where appropriate

```go
func TestDetectDrift(t *testing.T) {
    tests := []struct {
        name     string
        resource Resource
        changes  map[string]interface{}
        want     []AttributeDrift
    }{
        {
            name: "detect attribute change",
            resource: Resource{
                Attributes: map[string]interface{}{
                    "instance_type": "t2.micro",
                },
            },
            changes: map[string]interface{}{
                "instance_type": "t2.medium",
            },
            want: []AttributeDrift{
                {
                    Attribute: "instance_type",
                    OldValue:  "t2.micro",
                    NewValue:  "t2.medium",
                },
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := detectDrifts(tt.resource, tt.changes)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("detectDrifts() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

- Add integration tests for critical paths
- Use Docker containers for external dependencies
- Keep tests isolated and repeatable

### Multi-Cloud Testing (v0.5.0+)

When contributing multi-cloud features:

```go
// Test both AWS and GCP parsers
func TestMultiCloudParsing(t *testing.T) {
    // AWS test
    awsEvent := &outputs.Response{
        Source: "aws_cloudtrail",
        OutputFields: map[string]string{
            "aws.eventName": "ModifyInstanceAttribute",
        },
    }

    // GCP test
    gcpEvent := &outputs.Response{
        Source: "gcpaudit",
        OutputFields: map[string]string{
            "gcp.methodName": "compute.instances.setMetadata",
        },
    }

    // Test both parsers work correctly
}
```

- Test provider-specific functionality separately
- Test multi-provider scenarios (e.g., hybrid AWS+GCP deployments)
- Ensure configuration validation works for all providers
- Verify state backend selection (S3 for AWS, GCS for GCP)

## Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Build process or auxiliary tool changes

### Examples

```
feat(cloudtrail): add support for S3 event source

- Implement S3 bucket scanning for CloudTrail logs
- Add configuration options for S3 backend
- Update documentation

Closes #123
```

```
feat(gcp): add GCP Audit Logs support

- Implement GCP audit parser for Falco gcpaudit plugin
- Add GCS backend for Terraform state
- Support 100+ GCP events across 12 services
- Add comprehensive tests and documentation

Closes #150
```

```
fix(detector): handle nil resource attributes

Previously, the detector would panic when encountering resources
with nil attributes. Now it gracefully handles this case.

Fixes #456
```

## Pull Request Process

### Before Submitting

1. **Update your fork**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests**
   ```bash
   go test ./...
   golangci-lint run
   ```

3. **Update documentation** if needed

4. **Add tests** for new functionality

### Submitting

1. **Push to your fork**
   ```bash
   git push origin feature/my-feature
   ```

2. **Create Pull Request** on GitHub

3. **Fill out the PR template**

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback
3. Once approved, a maintainer will merge

### PR Guidelines

- Keep PRs focused and small
- One feature/fix per PR
- Write clear PR descriptions
- Link related issues
- Be responsive to feedback

## Community

### Where to Get Help

- 💬 **GitHub Discussions** - Ask questions and share ideas
- 🐛 **GitHub Issues** - Report bugs and request features
- 📧 **Email** - [keita.higaki@example.com](mailto:keita.higaki@example.com)
- 🐦 **Twitter** - [@keitahigaki](https://twitter.com/keitahigaki)

### Falco Community

Since TFDrift-Falco integrates with Falco, you might also find these resources helpful:

- **Falco Slack** - [Join #plugin-dev channel](https://slack.falco.org/)
- **Falco GitHub** - https://github.com/falcosecurity/falco

### Sysdig Community

- **Sysdig Community Slack** - https://sysdig.com/community/

## Recognition

Contributors will be recognized in:

- README.md contributors section
- Release notes
- Project website (coming soon)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to TFDrift-Falco!** 🎉
