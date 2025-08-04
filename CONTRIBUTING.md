# Contributing to Terraform Provider for Active Directory Groups

We welcome contributions to the Terraform Provider for Active Directory Groups! This document provides guidelines for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up the development environment
4. Create a topic branch for your changes
5. Make your changes
6. Test your changes
7. Submit a pull request

## Development Setup

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [golangci-lint](https://golangci-lint.run/usage/install/)
- Access to an Active Directory environment for testing

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/hknerts/terraform-provider-adgroups.git
   cd terraform-provider-adgroups
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the provider:
   ```bash
   make build
   ```

4. Run tests:
   ```bash
   make test
   ```

## Making Changes

### Branch Naming

Use descriptive branch names that indicate the type of change:

- `feature/add-user-resource` - for new features
- `bugfix/fix-group-creation` - for bug fixes
- `docs/update-readme` - for documentation changes
- `refactor/improve-client` - for refactoring

### Code Style

- Follow Go best practices and conventions
- Use `gofmt` to format your code
- Run `golangci-lint run` to check for issues
- Write clear, descriptive commit messages
- Keep functions and methods focused and small
- Add comments for complex logic

### Terraform Provider Conventions

- Follow [HashiCorp's provider development guidelines](https://developer.hashicorp.com/terraform/plugin/best-practices)
- Use the Terraform Plugin Framework
- Implement proper error handling and validation
- Use consistent naming conventions for resources and data sources
- Provide comprehensive documentation for all resources and data sources

### Adding New Resources

When adding a new resource:

1. Create the resource file in `internal/provider/`
2. Implement all required methods (Create, Read, Update, Delete)
3. Add proper schema validation
4. Include comprehensive tests
5. Update documentation
6. Add examples

### Adding New Data Sources

When adding a new data source:

1. Create the data source file in `internal/provider/`
2. Implement the Read method
3. Add proper schema validation
4. Include comprehensive tests
5. Update documentation
6. Add examples

## Testing

### Unit Tests

Run unit tests with:

```bash
make test
```

### Acceptance Tests

Acceptance tests require a real Active Directory environment. Set up environment variables:

```bash
export AD_SERVER="your-ad-server.com"
export AD_PORT="389"
export AD_BASE_DN="DC=example,DC=com"
export AD_USERNAME="CN=terraform,OU=ServiceAccounts,DC=example,DC=com"
export AD_PASSWORD="your-password"
export TF_ACC=1
```

Run acceptance tests:

```bash
make testacc
```

### Test Guidelines

- Write tests for all new functionality
- Use table-driven tests where appropriate
- Mock external dependencies when possible
- Include both positive and negative test cases
- Test edge cases and error conditions

### Test Structure

```go
func TestResourceGroup_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccResourceGroupConfig_basic(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("adgroups_group.test", "cn", "test-group"),
                    resource.TestCheckResourceAttrSet("adgroups_group.test", "dn"),
                ),
            },
        },
    })
}
```

## Documentation

### Code Documentation

- Add godoc comments for all exported functions and types
- Include examples in comments where helpful
- Document complex algorithms or business logic

### User Documentation

- Update `README.md` for new features
- Add or update resource/data source documentation in `docs/`
- Include examples for new functionality
- Update changelog with your changes

### Documentation Generation

Generate documentation using:

```bash
make docs
```

## Submitting Changes

### Pull Request Process

1. Update documentation and examples
2. Add or update tests for your changes
3. Ensure all tests pass
4. Run linting and formatting tools
5. Create a pull request with a clear description

### Pull Request Template

When creating a pull request, include:

- **Description**: What does this PR do?
- **Type of Change**: Bug fix, new feature, documentation, etc.
- **Testing**: How was this tested?
- **Checklist**: Complete the PR checklist
- **Related Issues**: Link to related issues

### Pull Request Checklist

- [ ] My code follows the project's code style
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- **Description**: Clear description of the issue
- **Steps to Reproduce**: Minimal steps to reproduce the behavior
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: OS, Terraform version, provider version
- **Configuration**: Relevant Terraform configuration (sanitized)
- **Logs**: Any relevant log output (sanitized)

### Feature Requests

When requesting features, include:

- **Use Case**: Why do you need this feature?
- **Proposed Solution**: How would you like it to work?
- **Alternatives**: Alternative solutions you've considered
- **Additional Context**: Any other relevant information

### Security Issues

For security-related issues, please email [security@example.com] instead of creating a public issue.

## Release Process

The maintainers handle releases using the following process:

1. Update CHANGELOG.md
2. Create a git tag
3. GitHub Actions builds and publishes the release
4. Update documentation if needed

## Getting Help

If you need help with development:

- Check existing documentation
- Look at similar implementations in the codebase
- Ask questions in pull requests or issues
- Reach out to maintainers

## Recognition

Contributors will be recognized in:

- CHANGELOG.md for significant changes
- README.md acknowledgments section
- GitHub contributors list

Thank you for contributing to the Terraform Provider for Active Directory Groups!