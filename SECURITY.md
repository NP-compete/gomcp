# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of gomcp seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please Do

- **Report the vulnerability privately** - Do not create a public GitHub issue for security vulnerabilities
- **Provide details** - Include as much information as possible about the vulnerability
- **Allow time** - Give us reasonable time to address the issue before any public disclosure

### Please Don't

- **Don't disclose publicly** - Please don't disclose the vulnerability publicly until we've had a chance to address it
- **Don't exploit** - Don't exploit the vulnerability beyond what's necessary to demonstrate it

### How to Report

1. **GitHub Security Advisories** (Preferred): Use [GitHub's security advisory feature](https://github.com/NP-compete/gomcp/security/advisories/new) to report vulnerabilities privately.

2. **Email**: If you prefer email, contact the maintainers directly (check the repository for contact information).

### What to Include

When reporting a vulnerability, please include:

- Type of vulnerability (e.g., injection, authentication bypass, etc.)
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability and how an attacker might exploit it

### Response Timeline

- **Initial Response**: Within 48 hours, we will acknowledge receipt of your report
- **Status Update**: Within 7 days, we will provide an initial assessment
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days

### After Reporting

1. We will confirm receipt of your vulnerability report
2. We will investigate and determine the impact
3. We will develop and test a fix
4. We will release a patch and publicly disclose the vulnerability (with credit to you, if desired)

## Security Best Practices for Users

When deploying gomcp in production:

### Authentication

- **Enable authentication** in production environments (`ENABLE_AUTH=true`)
- Use strong, unique session secrets (`SESSION_SECRET`)
- Configure OAuth/SSO with trusted identity providers

### Network Security

- Use HTTPS/TLS in production (configure `MCP_SSL_KEYFILE` and `MCP_SSL_CERTFILE`)
- Restrict CORS origins to trusted domains
- Use a reverse proxy (nginx, Caddy) for additional security headers

### Environment Variables

- Never commit `.env` files with secrets
- Use secret management solutions (Vault, AWS Secrets Manager, etc.)
- Rotate secrets regularly

### Container Security

- Run containers as non-root user (already configured in Dockerfile)
- Use read-only file systems where possible
- Scan images for vulnerabilities regularly

### Monitoring

- Enable logging and monitor for suspicious activity
- Set up alerts for authentication failures
- Review access logs regularly

## Known Security Considerations

### MCP Protocol

- The MCP protocol allows AI models to execute tools - ensure tools are properly sandboxed
- Validate all inputs from MCP clients
- Be cautious with tools that access the filesystem or execute commands

### Transport Security

- **stdio transport**: Only use for local, trusted clients (Claude Desktop)
- **HTTP/SSE transport**: Always use HTTPS in production
- Validate client certificates when possible

## Security Updates

Security updates will be released as patch versions (e.g., 0.1.1, 0.1.2). We recommend:

- Subscribing to GitHub releases for notifications
- Regularly updating to the latest patch version
- Reviewing the CHANGELOG for security-related fixes

## Acknowledgments

We appreciate the security research community's efforts in helping keep gomcp secure. Reporters of valid security issues will be acknowledged in our release notes (unless they prefer to remain anonymous).

---

Thank you for helping keep gomcp and its users safe!
