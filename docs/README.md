# Rosia CLI Documentation

This directory contains the Hugo-based documentation site for Rosia CLI.

## Local Development

### Prerequisites

- Hugo Extended v0.152.2 or later
- Git

### First Time Setup

Clone the theme (only needed for local development):

```bash
cd docs
git clone https://github.com/theNewDynamic/gohugo-theme-ananke.git themes/ananke
```

**Note:** The theme directory is excluded from git. GitHub Actions will automatically clone it during deployment.

### Running Locally

```bash
cd docs
hugo server -D
```

The site will be available at http://localhost:1313/

### Building

```bash
cd docs
hugo --gc --minify
```

The built site will be in `docs/public/`.

## Deployment

The documentation is automatically deployed to GitHub Pages at https://rosia.raucheacho.com/ when changes are pushed to the `main` branch.

## Structure

```
docs/
├── content/           # Markdown content files
│   ├── _index.md     # Homepage
│   ├── getting-started.md
│   ├── commands.md
│   ├── configuration.md
│   └── plugins.md
├── static/           # Static assets
│   └── CNAME        # Custom domain configuration
├── themes/          # Hugo themes
│   └── ananke/      # Current theme
└── hugo.toml        # Hugo configuration
```

## Adding Content

Create new pages:

```bash
hugo new content/page-name.md
```

Edit the frontmatter and content, then preview with `hugo server -D`.

## Theme

This site uses the [Ananke theme](https://github.com/theNewDynamic/gohugo-theme-ananke).

## Custom Domain

The site is configured to use the custom domain `rosia.raucheacho.com` via the CNAME file in `static/`.

To use a different domain:

1. Update `docs/static/CNAME` with your domain
2. Update `baseURL` in `docs/hugo.toml`
3. Configure DNS records to point to GitHub Pages
