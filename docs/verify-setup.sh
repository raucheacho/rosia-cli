#!/bin/bash

# Rosia Documentation Setup Verification Script

echo "🔍 Verifying Rosia Documentation Setup..."
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Hugo installation
echo "1. Checking Hugo installation..."
if command -v hugo &> /dev/null; then
    HUGO_VERSION=$(hugo version)
    echo -e "${GREEN}✓${NC} Hugo is installed: $HUGO_VERSION"
    
    if [[ $HUGO_VERSION == *"extended"* ]]; then
        echo -e "${GREEN}✓${NC} Hugo Extended is installed"
    else
        echo -e "${RED}✗${NC} Hugo Extended is required but not installed"
        echo "  Install with: brew install hugo"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} Hugo is not installed"
    echo "  Install with: brew install hugo"
    exit 1
fi
echo ""

# Check theme
echo "2. Checking theme..."
if [ -d "themes/ananke" ]; then
    echo -e "${GREEN}✓${NC} Ananke theme is installed"
else
    echo -e "${RED}✗${NC} Ananke theme is missing"
    echo "  Clone with: git clone https://github.com/theNewDynamic/gohugo-theme-ananke.git themes/ananke"
    exit 1
fi
echo ""

# Check content files
echo "3. Checking content files..."
CONTENT_FILES=(
    "content/_index.md"
    "content/getting-started.md"
    "content/commands.md"
    "content/configuration.md"
    "content/plugins.md"
    "content/faq.md"
)

for file in "${CONTENT_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $file exists"
    else
        echo -e "${RED}✗${NC} $file is missing"
        exit 1
    fi
done
echo ""

# Check CNAME
echo "4. Checking CNAME file..."
if [ -f "static/CNAME" ]; then
    DOMAIN=$(cat static/CNAME)
    echo -e "${GREEN}✓${NC} CNAME file exists: $DOMAIN"
    
    if [ "$DOMAIN" = "rosia.raucheacho.com" ]; then
        echo -e "${GREEN}✓${NC} Domain is correctly set"
    else
        echo -e "${YELLOW}⚠${NC} Domain is set to: $DOMAIN"
    fi
else
    echo -e "${RED}✗${NC} CNAME file is missing"
    exit 1
fi
echo ""

# Check hugo.toml
echo "5. Checking Hugo configuration..."
if [ -f "hugo.toml" ]; then
    echo -e "${GREEN}✓${NC} hugo.toml exists"
    
    if grep -q "rosia.raucheacho.com" hugo.toml; then
        echo -e "${GREEN}✓${NC} baseURL is correctly set"
    else
        echo -e "${YELLOW}⚠${NC} baseURL may not be correctly set"
    fi
    
    if grep -q "theme = 'ananke'" hugo.toml; then
        echo -e "${GREEN}✓${NC} Theme is configured"
    else
        echo -e "${RED}✗${NC} Theme is not configured"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} hugo.toml is missing"
    exit 1
fi
echo ""

# Test build
echo "6. Testing Hugo build..."
if hugo --gc --minify > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Hugo build successful"
    
    # Check output
    if [ -d "public" ]; then
        echo -e "${GREEN}✓${NC} public/ directory created"
        
        # Check CNAME in output
        if [ -f "public/CNAME" ]; then
            echo -e "${GREEN}✓${NC} CNAME file copied to public/"
        else
            echo -e "${RED}✗${NC} CNAME file not in public/"
            exit 1
        fi
        
        # Count pages
        PAGE_COUNT=$(find public -name "*.html" | wc -l | tr -d ' ')
        echo -e "${GREEN}✓${NC} Generated $PAGE_COUNT HTML pages"
    else
        echo -e "${RED}✗${NC} public/ directory not created"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} Hugo build failed"
    echo "  Run 'hugo --gc --minify' to see errors"
    exit 1
fi
echo ""

# Check GitHub Actions workflow
echo "7. Checking GitHub Actions workflow..."
WORKFLOW_FILE="../.github/workflows/deploy-docs.yml"
if [ -f "$WORKFLOW_FILE" ]; then
    echo -e "${GREEN}✓${NC} GitHub Actions workflow exists"
else
    echo -e "${RED}✗${NC} GitHub Actions workflow is missing"
    exit 1
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ All checks passed!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 Next steps:"
echo "  1. Configure DNS for rosia.raucheacho.com"
echo "  2. Push to GitHub: git push origin main"
echo "  3. Enable GitHub Pages in repository settings"
echo "  4. Wait for deployment and DNS propagation"
echo "  5. Visit https://rosia.raucheacho.com"
echo ""
echo "📚 Documentation:"
echo "  - Setup guide: GITHUB_PAGES_SETUP.md"
echo "  - Development: README.md"
echo "  - Summary: SUMMARY.md"
echo ""
echo "🚀 Local development:"
echo "  hugo server -D"
echo ""
