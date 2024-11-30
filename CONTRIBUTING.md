# Contributing to PasswordImmunity

Thank you for your interest in contributing to PasswordImmunity! This document provides guidelines and instructions for contributing to the project.

## Development Setup

1. Fork and clone the repository
2. Install dependencies:
   ```bash
   make deps
   ```
3. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Building and Testing

- Build the project: `make build`
- Run tests: `make test`
- Check code formatting: `make fmt`
- Run linter: `make lint`

## Pull Request Process

1. Ensure your code follows our formatting guidelines
2. Update documentation as needed
3. Add tests for new features
4. Ensure all tests pass
5. Update the README.md if needed
6. Create a Pull Request with a clear description

## Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and concise

## Commit Messages

- Use clear and descriptive commit messages
- Start with a verb (Add, Fix, Update, etc.)
- Reference issue numbers when applicable

## Testing

- Write unit tests for new features
- Maintain test coverage above 80%
- Include both positive and negative test cases

## Documentation

- Update API documentation for new endpoints
- Include examples in documentation
- Keep README.md current

## Questions or Problems?

- Open an issue for bugs
- Use discussions for questions
- Join our community chat for real-time help

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
