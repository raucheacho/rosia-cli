# Deployment Checklist for Rosia Documentation

Use this checklist to deploy the Rosia documentation to GitHub Pages with your custom domain.

## Pre-Deployment Verification

- [x] Hugo Extended installed (v0.152.2+)
- [x] Ananke theme installed in `themes/ananke/`
- [x] All content pages created (6 pages)
- [x] CNAME file exists in `static/CNAME`
- [x] Hugo config (`hugo.toml`) configured with custom domain
- [x] GitHub Actions workflow created (`.github/workflows/deploy-docs.yml`)
- [x] Local build successful (`hugo --gc --minify`)
- [x] Verification script passes (`./verify-setup.sh`)

## DNS Configuration

### Step 1: Choose DNS Method

**Option A: CNAME Record (Recommended)**
- [ ] Log in to your DNS provider (Cloudflare, Namecheap, GoDaddy, etc.)
- [ ] Navigate to DNS settings for `raucheacho.com`
- [ ] Add CNAME record:
  - Type: `CNAME`
  - Name: `rosia`
  - Value: `raucheacho.github.io` (replace with your GitHub username)
  - TTL: `3600` or Auto

**Option B: A Records**
- [ ] Add four A records pointing to:
  - `185.199.108.153`
  - `185.199.109.153`
  - `185.199.110.153`
  - `185.199.111.153`

### Step 2: Verify DNS Configuration

```bash
# Check CNAME
dig rosia.raucheacho.com CNAME

# Check A records
dig rosia.raucheacho.com A

# Or use online tool
# https://www.whatsmydns.net/#CNAME/rosia.raucheacho.com
```

- [ ] DNS records configured
- [ ] DNS propagation verified (may take 15-30 minutes)

## GitHub Repository Setup

### Step 3: Push Code to GitHub

```bash
# From project root
git add .
git commit -m "Add Hugo documentation site"
git push origin main
```

- [ ] Code pushed to GitHub
- [ ] All files committed (check `git status`)

### Step 4: Enable GitHub Pages

1. [ ] Go to repository on GitHub
2. [ ] Navigate to **Settings** → **Pages**
3. [ ] Under **Source**, select **GitHub Actions**
4. [ ] Save settings

### Step 5: Configure Custom Domain

1. [ ] In **Settings** → **Pages**
2. [ ] Under **Custom domain**, enter: `rosia.raucheacho.com`
3. [ ] Click **Save**
4. [ ] Wait for DNS check to complete (green checkmark)
5. [ ] Once verified, check **Enforce HTTPS**

## Deployment Verification

### Step 6: Monitor GitHub Actions

1. [ ] Go to **Actions** tab in repository
2. [ ] Find "Deploy Hugo Docs to GitHub Pages" workflow
3. [ ] Verify workflow runs successfully (green checkmark)
4. [ ] Check workflow logs if there are errors

### Step 7: Test Deployment

- [ ] Visit `https://rosia.raucheacho.com`
- [ ] Verify homepage loads correctly
- [ ] Test navigation menu (all links work)
- [ ] Check all pages:
  - [ ] Home
  - [ ] Getting Started
  - [ ] Commands
  - [ ] Configuration
  - [ ] Plugins
  - [ ] FAQ
- [ ] Verify HTTPS is working (padlock icon in browser)
- [ ] Test on mobile device (responsive design)

### Step 8: Verify SEO and Meta Tags

- [ ] Check page titles in browser tabs
- [ ] View page source and verify meta tags
- [ ] Test social media sharing (Open Graph tags)
- [ ] Submit sitemap to Google Search Console (optional)
  - Sitemap URL: `https://rosia.raucheacho.com/sitemap.xml`

## Post-Deployment

### Step 9: Update README

- [ ] Update main project README with documentation link
- [ ] Add badge for documentation site

Example badge:
```markdown
[![Documentation](https://img.shields.io/badge/docs-rosia.raucheacho.com-blue)](https://rosia.raucheacho.com)
```

### Step 10: Announce

- [ ] Update CHANGELOG.md with documentation site
- [ ] Announce on social media (optional)
- [ ] Update package managers (Homebrew, Scoop) with docs link

## Troubleshooting

### DNS Issues

**Problem:** Domain doesn't resolve
- [ ] Verify DNS records are correct
- [ ] Wait for DNS propagation (up to 48 hours)
- [ ] Clear DNS cache locally
- [ ] Use different DNS checker tool

**Commands:**
```bash
# macOS
sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder

# Linux
sudo systemd-resolve --flush-caches

# Windows
ipconfig /flushdns
```

### GitHub Actions Issues

**Problem:** Workflow fails
- [ ] Check workflow logs in Actions tab
- [ ] Verify Hugo version in workflow matches local
- [ ] Test build locally: `hugo --gc --minify`
- [ ] Check for syntax errors in content files

### HTTPS Issues

**Problem:** HTTPS not working
- [ ] Wait 24 hours for certificate provisioning
- [ ] Verify DNS is correctly configured
- [ ] Try removing and re-adding custom domain
- [ ] Check GitHub Pages status page

### 404 Errors

**Problem:** Pages show 404
- [ ] Verify CNAME file is in `static/` directory
- [ ] Check baseURL in `hugo.toml`
- [ ] Rebuild site: `hugo --gc --minify`
- [ ] Check GitHub Actions logs

## Maintenance Checklist

### Regular Updates

- [ ] Update content as needed
- [ ] Keep Hugo version updated
- [ ] Update theme periodically
- [ ] Monitor GitHub Actions for failures
- [ ] Check site performance (PageSpeed Insights)
- [ ] Review and update documentation quarterly

### Content Updates

To update documentation:

1. [ ] Edit files in `docs/content/`
2. [ ] Test locally: `hugo server -D`
3. [ ] Commit changes: `git commit -m "Update docs"`
4. [ ] Push to GitHub: `git push origin main`
5. [ ] Verify deployment in Actions tab
6. [ ] Check live site

## Success Criteria

All items below should be checked:

- [x] Hugo site created with all content pages
- [x] Theme installed and configured
- [x] Custom domain configured in CNAME and hugo.toml
- [x] GitHub Actions workflow created
- [x] Local build successful
- [ ] DNS configured and propagated
- [ ] Code pushed to GitHub
- [ ] GitHub Pages enabled
- [ ] Custom domain configured in GitHub
- [ ] Workflow runs successfully
- [ ] Site accessible at https://rosia.raucheacho.com
- [ ] HTTPS enabled
- [ ] All pages load correctly
- [ ] Navigation works
- [ ] Mobile responsive

## Quick Reference

### Local Development
```bash
cd docs
hugo server -D
# Visit: http://localhost:1313/
```

### Build Production
```bash
cd docs
hugo --gc --minify
```

### Verify Setup
```bash
cd docs
./verify-setup.sh
```

### Deploy
```bash
git add .
git commit -m "Update documentation"
git push origin main
```

## Support Resources

- **Hugo Docs**: https://gohugo.io/documentation/
- **GitHub Pages**: https://docs.github.com/en/pages
- **Ananke Theme**: https://github.com/theNewDynamic/gohugo-theme-ananke
- **Setup Guide**: `GITHUB_PAGES_SETUP.md`
- **Development Guide**: `README.md`

## Notes

- DNS propagation can take up to 48 hours (usually 15-30 minutes)
- HTTPS certificate provisioning can take up to 24 hours
- GitHub Actions workflow runs automatically on push to main
- Site rebuilds automatically when docs/ files change

---

**Deployment Date**: _____________
**Deployed By**: _____________
**Site URL**: https://rosia.raucheacho.com
**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete
