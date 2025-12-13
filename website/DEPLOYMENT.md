# Deployment Guide

This guide covers deploying the TFDrift-Falco website to Vercel.

## Quick Deploy

The fastest way to deploy is using the Vercel CLI:

### 1. Install Vercel CLI

```bash
npm i -g vercel
```

### 2. Deploy from the website directory

```bash
cd website
vercel
```

Follow the prompts to link your project and deploy.

### 3. Deploy to Production

```bash
vercel --prod
```

## Manual Deployment via Vercel Dashboard

### Step 1: Push to GitHub

1. Commit your changes:
   ```bash
   git add .
   git commit -m "Add website"
   git push origin main
   ```

### Step 2: Import to Vercel

1. Go to [vercel.com](https://vercel.com)
2. Click **Add New** → **Project**
3. Import your GitHub repository
4. Configure project:
   - **Framework Preset**: Next.js
   - **Root Directory**: `website`
   - **Build Command**: `npm run build`
   - **Output Directory**: `.next`

### Step 3: Configure Environment Variables

Add the following environment variables in Vercel Dashboard:

```bash
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX  # Your Google Analytics ID
NEXT_PUBLIC_SITE_URL=https://yourdomain.com  # Your site URL
```

### Step 4: Deploy

1. Click **Deploy**
2. Wait for the build to complete (usually 1-2 minutes)
3. Your site will be live at `your-project.vercel.app`

## Custom Domain Setup

To use a custom domain (e.g., `tfdrift-falco.com`), see the detailed guide:

👉 **[Custom Domain Setup Guide](./DOMAIN_SETUP.md)**

## Environment Variables

### Required Variables

None - the site works without environment variables.

### Optional Variables

- `NEXT_PUBLIC_GA_ID`: Google Analytics 4 Measurement ID
  - Format: `G-XXXXXXXXXX`
  - Get it from: [Google Analytics](https://analytics.google.com/)

- `NEXT_PUBLIC_SITE_URL`: Your production URL
  - Used for: SEO metadata, canonical URLs
  - Example: `https://tfdrift-falco.com`

### Setting Environment Variables

#### Via Vercel Dashboard

1. Go to **Settings** → **Environment Variables**
2. Add each variable with its value
3. Select environments: Production, Preview, Development
4. Click **Save**
5. Redeploy to apply changes

#### Via Vercel CLI

```bash
vercel env add NEXT_PUBLIC_GA_ID
# Paste your value when prompted

vercel env add NEXT_PUBLIC_SITE_URL
# Paste your value when prompted
```

#### Via .env.local (Development Only)

```bash
# website/.env.local
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

⚠️ **Never commit `.env.local` to Git!**

## Automatic Deployments

Vercel automatically deploys when you push to GitHub:

- **Production**: Deploys when you push to `main` branch
- **Preview**: Deploys for every pull request
- **Branch**: Deploy other branches via Vercel settings

### Configure Auto-Deploy

1. In Vercel Dashboard, go to **Settings** → **Git**
2. Ensure **Production Branch** is set to `main`
3. Enable **Automatic Deployments**

## Build Configuration

The website uses Next.js 14 with the App Router. Build configuration is in `next.config.js`:

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
}

module.exports = nextConfig
```

## Monitoring Deployments

### Via Vercel Dashboard

1. Go to your project
2. Click **Deployments** tab
3. View build logs, deployment status, and errors

### Via Vercel CLI

```bash
# List recent deployments
vercel ls

# View deployment logs
vercel logs [deployment-url]
```

## Rollback

If a deployment has issues:

1. Go to **Deployments** in Vercel Dashboard
2. Find a previous working deployment
3. Click **⋯** → **Promote to Production**

## Performance Optimization

The website is pre-optimized with:

- ✅ Static page generation where possible
- ✅ Automatic image optimization
- ✅ Code splitting and lazy loading
- ✅ CDN delivery via Vercel Edge Network
- ✅ Automatic HTTPS and HTTP/2

## Troubleshooting

### Build Fails

1. Check build logs in Vercel Dashboard
2. Verify all dependencies are in `package.json`
3. Test build locally:
   ```bash
   cd website
   npm run build
   ```

### Environment Variables Not Working

1. Ensure variables start with `NEXT_PUBLIC_` for client-side access
2. Redeploy after adding variables
3. Clear cache: Settings → Clear Build Cache & Redeploy

### Blog Posts Not Showing

1. Verify `content/blog/` directory exists
2. Check MDX files have correct frontmatter
3. Ensure `gray-matter` is installed:
   ```bash
   npm install gray-matter
   ```

### Documentation Redirect Not Working

The `/docs` page redirects to MkDocs on GitHub Pages. If it fails:

1. Verify MkDocs is deployed: https://higakikeita.github.io/tfdrift-falco/
2. Check `app/docs/page.tsx` has correct redirect URL

## GitHub Actions CI/CD (Optional)

For more control, you can set up GitHub Actions:

```yaml
# .github/workflows/deploy.yml
name: Deploy Website

on:
  push:
    branches: [main]
    paths:
      - 'website/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install dependencies
        working-directory: ./website
        run: npm ci
      - name: Build
        working-directory: ./website
        run: npm run build
      - name: Deploy to Vercel
        working-directory: ./website
        run: npx vercel --prod --token=${{ secrets.VERCEL_TOKEN }}
        env:
          VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
          VERCEL_ORG_ID: ${{ secrets.VERCEL_ORG_ID }}
          VERCEL_PROJECT_ID: ${{ secrets.VERCEL_PROJECT_ID }}
```

## Additional Resources

- [Vercel Documentation](https://vercel.com/docs)
- [Next.js Deployment](https://nextjs.org/docs/deployment)
- [Custom Domain Setup](./DOMAIN_SETUP.md)
- [Vercel CLI Reference](https://vercel.com/docs/cli)

## Support

If you encounter deployment issues:

1. Check [Vercel Status](https://www.vercel-status.com/)
2. Visit [Vercel Community](https://github.com/vercel/vercel/discussions)
3. Open an issue in [TFDrift-Falco repository](https://github.com/higakikeita/tfdrift-falco/issues)
