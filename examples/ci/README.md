# CI/CD Integration Examples

This directory contains CI/CD pipeline examples for integrating GODNSLOG OAST scanning into your continuous integration workflows.

## Supported Platforms

- GitHub Actions
- GitLab CI/CD
- Jenkins

## Prerequisites

1. Deploy GODNSLOG server
2. Create API key with appropriate scopes
3. Configure secrets in your CI/CD platform:
   - `GODNSLOG_API_URL`: Your GODNSLOG server URL
   - `GODNSLOG_API_KEY`: Your GODNSLOG API key

## GitHub Actions

### Setup

1. Add secrets to your repository:
   - Settings > Secrets and variables > Actions
   - Add `GODNSLOG_API_URL`
   - Add `GODNSLOG_API_KEY`

2. Copy `github-actions.yml` to `.github/workflows/oast-scan.yml`

3. Commit and push

### Features

- Automatic case creation for each scan
- Payload generation with 24h expiration
- Nuclei integration for automated scanning
- Real-time interaction polling
- High-risk detection as pipeline gate
- Markdown report export
- Summary posted to PR/commit

### Pipeline Gate

The pipeline will fail if high-risk interactions are detected:
- Cloud metadata access (169.254.169.254)
- Internal network access
- SSRF patterns

## GitLab CI/CD

### Setup

1. Add CI/CD variables:
   - Settings > CI/CD > Variables
   - Add `GODNSLOG_API_URL` (masked)
   - Add `GODNSLOG_API_KEY` (masked)

2. Copy `gitlab-ci.yml` to `.gitlab-ci.yml`

3. Commit and push

### Features

- Same as GitHub Actions
- GitLab artifact reports
- Merge request integration

## Jenkins

### Setup

1. Configure credentials:
   - Manage Jenkins > Credentials > System > Global credentials
   - Add `godnslog-api-url` (Secret text)
   - Add `godnslog-api-key` (Secret text)

2. Copy `jenkinsfile` to `Jenkinsfile`

3. Configure pipeline job to use the Jenkinsfile

### Features

- Jenkins pipeline support
- HTML report publishing
- Build status integration

## Customization

### Modify Scan Targets

Edit the `-u` parameter in Nuclei command:
```yaml
-u https://your-target.com
```

### Add Custom Templates

Add your own Nuclei templates:
```bash
mkdir oast-templates
# Add your custom templates
```

### Adjust Timeout

Modify polling timeout:
```bash
godnslog interaction poll $PAYLOAD --timeout 30m
```

### Change Payload Template

Modify payload generation:
```bash
godnslog payload create \
  --template "{{.Token}}.custom-domain.com" \
  --case-id $CASE_ID
```

## Best Practices

1. **API Key Management**: Use scoped API keys with minimal permissions
2. **Case Organization**: Use descriptive case titles and tags
3. **Payload Expiration**: Set appropriate expiration times (24h for CI/CD)
4. **Noise Reduction**: Filter out known safe patterns
5. **Report Review**: Always review generated reports manually
6. **Pipeline Gates**: Use high-risk detection as quality gates

## Troubleshooting

### API Connection Issues

- Verify API URL is accessible from CI runner
- Check API key has correct permissions
- Review CI logs for authentication errors

### No Interactions Detected

- Verify Nuclei templates are configured correctly
- Check target is reachable from CI environment
- Ensure payload is properly inserted into requests

### High-Risk False Positives

- Adjust high-risk detection patterns
- Add whitelist for known safe domains
- Review and adjust pipeline gate logic
