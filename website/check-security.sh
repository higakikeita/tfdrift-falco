#!/bin/bash

echo "🔍 TFDrift-Falco Website Security Check"
echo "========================================"
echo ""

# Get React version
REACT_VERSION=$(node -p "require('./package.json').dependencies.react" 2>/dev/null)
NEXT_VERSION=$(node -p "require('./package.json').dependencies.next" 2>/dev/null)

echo "📦 Current Versions:"
echo "  React: $REACT_VERSION"
echo "  Next.js: $NEXT_VERSION"
echo ""

# Check React version for CVE-2025-55182, CVE-2025-55183, CVE-2025-55184, CVE-2025-67779
echo "🔐 Checking for known vulnerabilities..."
echo ""

VULNERABLE=false

# Extract version number (remove ^ or ~ if present)
REACT_VER_NUM=$(echo $REACT_VERSION | sed 's/[\^~]//g')

# Check React version
if [[ "$REACT_VER_NUM" == "19.2.1" ]] || [[ "$REACT_VER_NUM" == "19.2.2" ]] || \
   [[ "$REACT_VER_NUM" == "19.1.3" ]] || [[ "$REACT_VER_NUM" == "19.0.2" ]]; then
    echo "⚠️  VULNERABLE: React $REACT_VER_NUM"
    echo "   CVE-2025-55182: Remote Code Execution (Critical - CVSS 10.0)"
    echo "   CVE-2025-55183: Source Code Exposure (Medium - CVSS 5.3)"
    echo "   CVE-2025-55184: Denial of Service (High - CVSS 7.5)"
    echo "   CVE-2025-67779: Incomplete fix for CVE-2025-55184 (High - CVSS 7.5)"
    echo ""
    echo "   🔧 Fix: npm install react@19.2.3 react-dom@19.2.3"
    echo ""
    VULNERABLE=true
elif [[ "$REACT_VER_NUM" == "19.2.3" ]] || [[ "$REACT_VER_NUM" == "19.1.4" ]] || [[ "$REACT_VER_NUM" == "19.0.3" ]]; then
    echo "✅ SAFE: React $REACT_VER_NUM (patched)"
else
    echo "ℹ️  React $REACT_VER_NUM - Please verify against latest security advisories"
fi

echo ""

# Check Next.js version
NEXT_VER_NUM=$(echo $NEXT_VERSION | sed 's/[\^~]//g')

if [[ "$NEXT_VER_NUM" == "16.0.6" ]] || [[ "$NEXT_VER_NUM" < "15.0.5" ]]; then
    echo "⚠️  VULNERABLE: Next.js $NEXT_VER_NUM"
    echo "   CVE-2025-66478: Related to React Server Components"
    echo ""
    echo "   🔧 Fix: npm install next@latest"
    echo ""
    VULNERABLE=true
else
    echo "✅ SAFE: Next.js $NEXT_VER_NUM"
fi

echo ""
echo "========================================"

if [ "$VULNERABLE" = true ]; then
    echo "❌ SECURITY ISSUES FOUND"
    echo ""
    echo "Run the following commands to fix:"
    echo "  cd /Users/keita.higaki/tfdrift-falco/website"
    echo "  npm install react@19.2.3 react-dom@19.2.3"
    echo "  npm run build"
    echo "  git add package.json package-lock.json"
    echo "  git commit -m 'security: Update React to 19.2.3 (CVE-2025-55182/55183/55184/67779)'"
    echo ""
    exit 1
else
    echo "✅ No known vulnerabilities detected"
    echo ""
    echo "💡 Tip: Run 'npm audit' and 'snyk test' for additional checks"
    exit 0
fi
