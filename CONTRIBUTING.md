# Contributing to Flixsrota

Thank you for your interest in contributing to Flixsrota! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- FFmpeg installed on your system
- Redis (optional, for queue testing)
- Protobuf compiler (for development)

### Development Setup

1. **Fork and clone the repository:**
```bash
git clone https://github.com/nikhil0verma/flixsrota.git
cd flixsrota
```

2. **Install dependencies:**
```bash
make deps
```

3. **Generate protobuf code:**
```bash
make proto
```

4. **Run tests:**
```bash
make test
```

## 🛠 Development Workflow

### Code Style

- Follow Go best practices and conventions
- Use `gofmt` for code formatting
- Write clear, descriptive commit messages
- Add tests for new functionality

### Making Changes

1. **Create a feature branch:**
```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes and test:**
```bash
make test
make lint
```

3. **Commit your changes:**
```bash
git commit -m "feat: add new feature description"
```

4. **Push and create a pull request:**
```bash
git push origin feature/your-feature-name
```

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/queue -v
```

### Writing Tests

- Write unit tests for new functionality
- Aim for good test coverage
- Use descriptive test names
- Mock external dependencies

## 📝 Documentation

### Code Documentation

- Add comments for exported functions and types
- Follow Go documentation conventions
- Update README.md for user-facing changes

### API Documentation

- Update protobuf definitions when adding new services
- Document new configuration options
- Add examples for new features

## 🔧 Adding New Features

### Queue Adapters

1. Implement the `Queue` interface in `internal/queue/`
2. Add configuration options in `internal/config/config.go`
3. Register the adapter in `internal/core/server.go`
4. Add tests for the new adapter

### Storage Adapters

1. Implement the `Storage` interface in `internal/storage/`
2. Add configuration options in `internal/config/config.go`
3. Register the adapter in `internal/core/server.go`
4. Add tests for the new adapter

### gRPC Services

1. Define the service in `proto/flixsrota.proto`
2. Generate protobuf code: `make proto`
3. Implement the service in `internal/grpc/`
4. Add tests for the new service

## 🐛 Bug Reports

When reporting bugs, please include:

- **Description:** Clear description of the issue
- **Steps to reproduce:** Detailed steps to reproduce the bug
- **Expected behavior:** What you expected to happen
- **Actual behavior:** What actually happened
- **Environment:** OS, Go version, FFmpeg version
- **Logs:** Relevant log output

## 💡 Feature Requests

When requesting features, please include:

- **Description:** Clear description of the feature
- **Use case:** Why this feature is needed
- **Proposed implementation:** How you think it should work
- **Alternatives considered:** Other approaches you considered

## 🤝 Pull Request Guidelines

### Before Submitting

- [ ] Code follows Go conventions
- [ ] Tests pass: `make test`
- [ ] Linting passes: `make lint`
- [ ] Documentation is updated
- [ ] Commit messages follow conventional format

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

## 📋 Code Review Process

1. **Automated checks** must pass
2. **Code review** by maintainers
3. **Address feedback** and make requested changes
4. **Merge** when approved

## 🏷 Release Process

### Creating a Release

1. **Update version:**
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. **Build releases:**
```bash
make build-all
make release
```

3. **Create GitHub release** with release notes

## 📞 Getting Help

- **Issues:** [GitHub Issues](https://github.com/nikhil0verma/flixsrota/issues)
- **Discussions:** [GitHub Discussions](https://github.com/nikhil0verma/flixsrota/discussions)
- **Documentation:** [Wiki](https://github.com/nikhil0verma/flixsrota/wiki)

## 📄 License

By contributing to Flixsrota, you agree that your contributions will be licensed under the Apache License 2.0.

## 🙏 Acknowledgments

Thank you for contributing to Flixsrota! Your contributions help make this project better for everyone. 