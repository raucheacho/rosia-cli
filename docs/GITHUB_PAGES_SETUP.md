# GitHub Pages Setup Guide

This guide will help you deploy the Rosia documentation to GitHub Pages with your custom domain `rosia.raucheacho.com`.

## Prerequisites

- GitHub repository with the Rosia CLI code
- Custom domain `rosia.raucheacho.com`
- Access to your domain's DNS settings

## Step 1: Enable GitHub Pages

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Pages**
3. Under **Source**, select **GitHub Actions**
4. Save the settings

## Step 2: Configure DNS

Add the following DNS records for your domain `rosia.raucheacho.com`:

### Option A: Using CNAME (Recommended)

Add a CNAME record:

```
Type: CNAME
Name: rosia
Value: <your-github-username>.github.io
TTL: 3600 (or your provider's default)
```

For example, if your GitHub username is `raucheacho`:
```
Type: CNAME
Name: rosia
Value: raucheacho.github.io
```

### Option B: Using A Records

If you prefer A records, add these four records:

```
Type: A
Name: rosia
Value: 185.199.108.153
TTL: 3600

Type: A
Name: rosia
Value: 185.199.109.153
TTL: 3600

Type: A
Name: rosia
Value: 185.199.110.153
TTL: 3600

Type: A
Name: rosia
Value: 185.199.111.153
TTL: 3600
```

## Step 3: Verify DNS Propagation

Wait for DNS propagation (can take up to 48 hours, but usually 15-30 minutes).

Check DNS propagation:

```bash
# Check CNAME
dig rosia.raucheacho.com CNAME

# Check A records
dig rosia.raucheacho.com A

# Or use online tools
# https://www.whatsmydns.net/#CNAME/rosia.raucheacho.com
```

## Step 4: Push to GitHub

The GitHub Actions workflow is already configured in `.github/workflows/deploy-docs.yml`.

**Note:** The theme directory (`docs/themes/`) is excluded from git (via `.gitignore`). The GitHub Actions workflow will automatically clone the theme during deployment.

Push your changes:

```bash
git add .
git commit -m "Add Hugo documentation site"
git push origin main
```

## Step 5: Monitor Deployment

1. Go to your repository on GitHub
2. Click on the **Actions** tab
3. Watch the "Deploy Hugo Docs to GitHub Pages" workflow
4. Wait for it to complete (usually 1-2 minutes)

## Step 6: Configure Custom Domain in GitHub

1. Go to **Settings** → **Pages**
2. Under **Custom domain**, enter: `rosia.raucheacho.com`
3. Click **Save**
4. Wait for DNS check to complete
5. Once verified, check **Enforce HTTPS**

## Step 7: Verify Deployment

Visit your site:

```
https://rosia.raucheacho.com
```

You should see the Rosia CLI documentation homepage.

## Troubleshooting

### DNS Not Resolving

**Problem:** `rosia.raucheacho.com` doesn't resolve

**Solution:**
1. Verify DNS records are correct
2. Wait for DNS propagation (up to 48 hours)
3. Clear your DNS cache:
   ```bash
   # macOS
   sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder
   
   # Linux
   sudo systemd-resolve --flush-caches
   
   # Windows
   ipconfig /flushdns
   ```

### 404 Error

**Problem:** Site shows 404 error

**Solution:**
1. Check that GitHub Actions workflow completed successfully
2. Verify the CNAME file exists in `docs/static/CNAME`
3. Rebuild the site:
   ```bash
   cd docs
   hugo --gc --minify
   git add .
   git commit -m "Rebuild site"
   git push
   ```

### HTTPS Not Working

**Problem:** HTTPS certificate not provisioning

**Solution:**
1. Wait 24 hours for GitHub to provision the certificate
2. Ensure DNS is correctly configured
3. Try removing and re-adding the custom domain in GitHub Pages settings

### Workflow Failing

**Problem:** GitHub Actions workflow fails

**Solution:**
1. Check the workflow logs in the Actions tab
2. Verify Hugo version in `.github/workflows/deploy-docs.yml` matches your local version
3. Ensure all content files have valid frontmatter
4. Test build locally:
   ```bash
   cd docs
   hugo --gc --minify
   ```

### Theme Not Loading

**Problem:** Site loads but styling is broken

**Solution:**
1. Verify the theme is correctly cloned in `docs/themes/ananke`
2. Check `baseURL` in `docs/hugo.toml` matches your domain
3. Rebuild and redeploy

## Updating Documentation

To update the documentation:

1. Edit content files in `docs/content/`
2. Test locally:
   ```bash
   cd docs
   hugo server -D
   ```
3. Commit and push:
   ```bash
   git add docs/
   git commit -m "Update documentation"
   git push origin main
   ```

The site will automatically rebuild and deploy via GitHub Actions.

## Local Development

### Start Development Server

```bash
cd docs
hugo server -D
```

Visit http://localhost:1313/

### Build Production Site

```bash
cd docs
hugo --gc --minify
```

Output will be in `docs/public/`

### Test Production Build Locally

```bash
cd docs/public
python3 -m http.server 8080
```

Visit http://localhost:8080/

## Custom Domain Configuration

The custom domain is configured in two places:

1. **CNAME file**: `docs/static/CNAME`
   ```
   rosia.raucheacho.com
   ```

2. **Hugo config**: `docs/hugo.toml`
   ```toml
   baseURL = 'https://rosia.raucheacho.com/'
   ```

If you change your domain, update both files.

## GitHub Actions Workflow

The workflow (`.github/workflows/deploy-docs.yml`) automatically:

1. Checks out the repository
2. Installs Hugo Extended
3. Builds the site with `hugo --gc --minify`
4. Uploads the build artifact
5. Deploys to GitHub Pages

The workflow runs on:
- Push to `main` branch (when `docs/**` files change)
- Manual trigger via workflow_dispatch

## DNS Provider Examples

### Cloudflare

1. Log in to Cloudflare
2. Select your domain `raucheacho.com`
3. Go to **DNS** → **Records**
4. Add CNAME record:
   - Type: CNAME
   - Name: rosia
   - Target: raucheacho.github.io
   - Proxy status: DNS only (gray cloud)
   - TTL: Auto

### Namecheap

1. Log in to Namecheap
2. Go to **Domain List** → **Manage**
3. Select **Advanced DNS**
4. Add CNAME record:
   - Type: CNAME Record
   - Host: rosia
   - Value: raucheacho.github.io
   - TTL: Automatic

### GoDaddy

1. Log in to GoDaddy
2. Go to **My Products** → **DNS**
3. Add CNAME record:
   - Type: CNAME
   - Name: rosia
   - Value: raucheacho.github.io
   - TTL: 1 Hour

### Google Domains

1. Log in to Google Domains
2. Select your domain
3. Go to **DNS**
4. Add custom resource record:
   - Name: rosia
   - Type: CNAME
   - TTL: 1H
   - Data: raucheacho.github.io

## Security

### HTTPS

GitHub Pages automatically provisions and renews SSL certificates for custom domains using Let's Encrypt.

To enable HTTPS:
1. Wait for DNS to propagate
2. Go to **Settings** → **Pages**
3. Check **Enforce HTTPS**

### Content Security

The documentation is static HTML/CSS/JS with no server-side code, making it inherently secure.

## Performance

### Optimization

The Hugo build process includes:
- Minification of HTML, CSS, and JS
- Garbage collection of unused resources
- Asset fingerprinting for cache busting

### CDN

GitHub Pages uses a global CDN (Fastly) for fast content delivery worldwide.

### Monitoring

Monitor your site's performance:
- [Google PageSpeed Insights](https://pagespeed.web.dev/)
- [GTmetrix](https://gtmetrix.com/)
- [WebPageTest](https://www.webpagetest.org/)

## Maintenance

### Regular Updates

1. Keep Hugo version updated in workflow
2. Update theme periodically:
   ```bash
   cd docs/themes/ananke
   git pull origin main
   ```
3. Review and update content regularly

### Backup

The documentation is version-controlled in Git, providing automatic backup and history.

## Support

If you encounter issues:

1. Check [Hugo Documentation](https://gohugo.io/documentation/)
2. Review [GitHub Pages Documentation](https://docs.github.com/en/pages)
3. Check [GitHub Actions logs](https://github.com/raucheacho/rosia-cli/actions)
4. Open an issue on the repository

## Additional Resources

- [Hugo Quick Start](https://gohugo.io/getting-started/quick-start/)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Ananke Theme Documentation](https://github.com/theNewDynamic/gohugo-theme-ananke)
- [Custom Domain Configuration](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site)
