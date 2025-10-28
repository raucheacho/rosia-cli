# Quick Start Guide

## 🚀 Deploy in 5 Minutes

### 1. Configure DNS (2 minutes)

Add CNAME record to your DNS:
```
Type: CNAME
Name: rosia
Value: raucheacho.github.io
```

### 2. Push to GitHub (1 minute)

```bash
git add .
git commit -m "Add Hugo documentation"
git push origin main
```

### 3. Enable GitHub Pages (1 minute)

1. Go to **Settings** → **Pages**
2. Source: **GitHub Actions**
3. Custom domain: `rosia.raucheacho.com`
4. Enable **Enforce HTTPS**

### 4. Wait & Verify (1 minute)

- Wait for GitHub Actions to complete
- Visit: https://rosia.raucheacho.com

Done! 🎉

---

## 💻 Local Development

### Start Server
```bash
cd docs
hugo server -D
```
Visit: http://localhost:1313/

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

---

## 📝 Update Content

### Edit Page
```bash
# Edit any file in docs/content/
nano docs/content/getting-started.md
```

### Preview Changes
```bash
cd docs
hugo server -D
```

### Deploy Changes
```bash
git add docs/
git commit -m "Update docs"
git push origin main
```

Auto-deploys via GitHub Actions!

---

## 🔧 Common Commands

| Task | Command |
|------|---------|
| Start dev server | `hugo server -D` |
| Build production | `hugo --gc --minify` |
| Verify setup | `./verify-setup.sh` |
| Create new page | `hugo new content/page.md` |
| Check Hugo version | `hugo version` |

---

## 📚 Documentation Files

| File | Purpose |
|------|---------|
| `QUICK_START.md` | This file - quick reference |
| `GITHUB_PAGES_SETUP.md` | Complete deployment guide |
| `DEPLOYMENT_CHECKLIST.md` | Step-by-step checklist |
| `README.md` | Local development guide |
| `SUMMARY.md` | Project overview |

---

## 🆘 Troubleshooting

### Site not loading?
1. Check DNS: `dig rosia.raucheacho.com`
2. Check GitHub Actions: Repository → Actions tab
3. Wait for DNS propagation (15-30 min)

### Build failing?
```bash
cd docs
hugo --gc --minify
# Check for errors
```

### Need help?
- Read `GITHUB_PAGES_SETUP.md`
- Check GitHub Actions logs
- Review Hugo docs: https://gohugo.io/

---

## ✅ Verification

Run this to verify everything:
```bash
cd docs
./verify-setup.sh
```

Should see: `✓ All checks passed!`

---

## 🌐 URLs

- **Live Site**: https://rosia.raucheacho.com
- **Local Dev**: http://localhost:1313/
- **GitHub Repo**: https://github.com/raucheacho/rosia-cli
- **GitHub Actions**: https://github.com/raucheacho/rosia-cli/actions

---

**Need more details?** See `GITHUB_PAGES_SETUP.md`
