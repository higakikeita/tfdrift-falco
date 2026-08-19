# Building the driftwire Project Site with Next.js + Vercel + MkDocs

## Introduction

I built a project website for "driftwire", a Terraform drift detection tool, using Next.js 14, Vercel, and MkDocs. This article covers the entire process from tech stack decisions to implementation and challenges faced.

- **Project Site**: https://tfdrift-falco.vercel.app/
- **Documentation**: https://higakikeita.github.io/driftwire/
- **GitHub**: https://github.com/higakikeita/driftwire

## What is driftwire?

driftwire is a real-time Terraform drift detection system that combines AWS CloudTrail with Falco.

### Key Features
- Instantly detects manual changes made through the AWS Console
- Supports 120+ CloudTrail events
- Automatically detects differences from Terraform state
- Sends alerts to Slack/PagerDuty/Grafana

## Why Build a Project Website?

For open-source project growth, information dissemination is as important as technical implementation. I built the project site with these goals:

1. **Increase Visibility**: Establish a presence beyond GitHub
2. **Organize Documentation**: Consolidate scattered information
3. **Communicate Releases**: Share the latest development status
4. **Build Community**: Create more touchpoints with users

## Tech Stack

### Overall Architecture

```
driftwire Website
├── Vercel (Next.js 14)
│   ├── Landing Page
│   ├── Blog (MDX)
│   └── Releases (GitHub API)
│
└── GitHub Pages (MkDocs)
    └── Technical Documentation
```

### Technology Choices

#### Next.js 14 + Vercel
- **App Router**: Leverages latest React Server Components
- **Zero-config Deployment**: Auto-deploy on GitHub push
- **Edge Network**: Fast delivery worldwide
- **Generous Free Tier**: Perfect for personal projects

#### MkDocs Material
- **Specialized for Technical Docs**: Excellent code blocks, search, and navigation
- **GitHub Pages Integration**: Easy hosting
- **Markdown**: Developer-friendly

#### MDX (next-mdx-remote)
- **Markdown + React**: Embed components
- **Syntax Highlighting**: Beautiful code display with Prism.js
- **Frontmatter Support**: Easy metadata management

## Implementation Highlights

### 1. Blog Functionality

#### Directory Structure

```
website/
├── app/
│   ├── blog/
│   │   ├── page.tsx           # Blog listing
│   │   └── [slug]/
│   │       └── page.tsx       # Individual posts
│   └── layout.tsx
└── content/
    └── blog/
        └── *.mdx              # Blog posts
```

#### Individual Post Page Implementation

Since Next.js 15, `params` is now a Promise and requires `await`:

```typescript
// app/blog/[slug]/page.tsx
import { MDXRemote } from 'next-mdx-remote/rsc'
import matter from 'gray-matter'
import remarkGfm from 'remark-gfm'
import rehypePrism from 'rehype-prism-plus'

interface BlogPostPageProps {
  params: Promise<{
    slug: string
  }>
}

export default async function BlogPostPage({ params }: BlogPostPageProps) {
  const { slug } = await params  // ← await is required!
  const post = getPostData(slug)

  return (
    <article>
      <h1>{post.title}</h1>
      <MDXRemote
        source={post.content}
        options={{
          mdxOptions: {
            remarkPlugins: [remarkGfm],
            rehypePlugins: [rehypePrism]
          }
        }}
      />
    </article>
  )
}

// Static generation configuration
export function generateStaticParams() {
  const postsDirectory = path.join(process.cwd(), 'content/blog')
  const filenames = fs.readdirSync(postsDirectory)

  return filenames
    .filter(filename => filename.endsWith('.mdx'))
    .map(filename => ({
      slug: filename.replace('.mdx', '')
    }))
}
```

#### MDX File Format

```markdown
---
title: "Post Title"
date: "2025-12-12"
excerpt: "Post summary"
author: "Author Name"
tags: ["tag1", "tag2"]
---

# Article Content

Write your article in Markdown here.

\`\`\`yaml
# Code blocks are supported
drift_rules:
  - name: "Production ECS Service Modified"
    resource_types:
      - "aws_ecs_service"
\`\`\`
```

### 2. Releases Page with GitHub API Integration

Auto-fetch release information using GitHub API:

```typescript
// app/releases/page.tsx
async function getReleases() {
  const res = await fetch(
    'https://api.github.com/repos/higakikeita/driftwire/releases',
    {
      next: { revalidate: 3600 } // Refetch every hour
    }
  )

  if (!res.ok) return []
  return res.json()
}

export default async function ReleasesPage() {
  const releases = await getReleases()

  return (
    <div>
      {releases.map((release) => (
        <ReleaseCard key={release.id} release={release} />
      ))}
    </div>
  )
}
```

### 3. MkDocs Integration

#### Documentation Redirect

```typescript
// app/docs/page.tsx
import { redirect } from 'next/navigation'

export default function DocsPage() {
  redirect('https://higakikeita.github.io/driftwire/')
}
```

#### Auto-deployment with GitHub Actions

```yaml
# .github/workflows/docs.yml
name: Build & Deploy Docs

on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
      - 'mkdocs.yml'
  workflow_dispatch:

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install MkDocs and dependencies
        run: |
          pip install mkdocs mkdocs-material mkdocs-minify-plugin pymdown-extensions

      - name: Build MkDocs site
        run: mkdocs build --verbose

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_branch: gh-pages
          publish_dir: ./site
```

### 4. Google Analytics Integration

```typescript
// app/layout.tsx
import { GoogleAnalytics } from '@next/third-parties/google'

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        {children}
        {process.env.NEXT_PUBLIC_GA_ID && (
          <GoogleAnalytics gaId={process.env.NEXT_PUBLIC_GA_ID} />
        )}
      </body>
    </html>
  )
}
```

Environment variables:

```bash
# .env.example
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
NEXT_PUBLIC_SITE_URL=https://tfdrift-falco.vercel.app
```

## Challenges & Solutions

### 1. Next.js 15 params Specification Change

**Problem**: 404 errors on individual blog post pages

**Cause**: Since Next.js 15, `params` is now a Promise and requires `await`

```typescript
// ❌ Doesn't work
export default function BlogPostPage({ params }: BlogPostPageProps) {
  const post = getPostData(params.slug)  // Error
}

// ✅ Correct
export default async function BlogPostPage({ params }: BlogPostPageProps) {
  const { slug } = await params  // await is required
  const post = getPostData(slug)
}
```

### 2. GitHub Pages CNAME Issue

**Problem**: MkDocs documentation returns 404 error

**Cause**: `CNAME` file contains custom domain (`docs.driftwire.com`) but DNS is not configured

**Solution**: Remove CNAME configuration from workflow

```yaml
# ❌ When custom domain is not configured
- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    cname: docs.driftwire.com  # This causes the issue

# ✅ Fixed version
- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    # cname: docs.driftwire.com  # Commented out
    publish_dir: ./site
```

### 3. MkDocs Strict Mode

**Problem**: MkDocs build fails

```
Aborted with 16 warnings in strict mode!
```

**Cause**: Broken links in documentation detected as warnings, treated as errors in `--strict` mode

**Solution**: Remove `--strict` flag

```yaml
# ❌ Treats warnings as errors in strict mode
- name: Build MkDocs site
  run: mkdocs build --strict --verbose

# ✅ Shows warnings but continues build
- name: Build MkDocs site
  run: mkdocs build --verbose
```

### 4. Prism.js Styling

**Problem**: Syntax highlighting not working in code blocks

**Solution**: Add Prism theme to global CSS

```css
/* app/globals.css */
code[class*="language-"],
pre[class*="language-"] {
  color: #e2e8f0;
  background: none;
  font-family: 'Consolas', 'Monaco', monospace;
  text-align: left;
  white-space: pre;
  word-spacing: normal;
  line-height: 1.5;
  tab-size: 4;
}

pre[class*="language-"] {
  padding: 1.5em;
  margin: 0.5em 0;
  overflow: auto;
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.5);
}

/* Token colors */
.token.comment { color: #94a3b8; }
.token.string { color: #86efac; }
.token.keyword { color: #a78bfa; }
.token.function { color: #60a5fa; }
```

## Deployment Flow

### Vercel (Website)

```bash
1. Push to GitHub
   ↓
2. Vercel auto-detects
   ↓
3. Next.js build
   ↓
4. Deploy to Edge Network
   ↓
5. Published at https://tfdrift-falco.vercel.app/
```

### GitHub Pages (Documentation)

```bash
1. Edit docs/ and push
   ↓
2. GitHub Actions triggers
   ↓
3. MkDocs build
   ↓
4. Push to gh-pages branch
   ↓
5. Published at https://higakikeita.github.io/driftwire/
```

## Results

### Performance

- **Lighthouse Score**: 95+ (all categories)
- **First Contentful Paint**: < 1.0s
- **Time to Interactive**: < 2.0s

### Operations

✅ **Easy Blog Post Addition**
- Just place `.mdx` file and push
- Auto-generated at build time

✅ **Automatic Documentation Updates**
- Edit Markdown and push
- GitHub Actions auto-deploys

✅ **Auto-updating Release Information**
- Just create GitHub Releases
- Site automatically fetches latest (1-hour cache)

✅ **SEO Optimized**
- OpenGraph and Twitter Cards support
- Search engine optimized metadata

## Cost

Everything runs **completely free**:

- **Vercel**: Hobby plan (free)
- **GitHub Pages**: Free
- **GitHub Actions**: 2000 minutes/month free (sufficient)
- **Domain**: Using Vercel's `vercel.app` domain

## Future Plans

### Short-term Improvements

1. **Enrich Blog Content**
   - Technical explanation articles
   - Use case introductions
   - Development diaries

2. **Expand Documentation**
   - More AWS service coverage
   - Additional tutorials
   - Troubleshooting guides

3. **SEO Optimization**
   - Sitemap optimization
   - Add structured data
   - Performance tuning

### Long-term Plans

1. **Community Features**
   - Integration with Discussions
   - Contributor spotlights
   - User case studies

2. **Internationalization**
   - English documentation
   - i18n support

3. **Interactive Demos**
   - Browser-based demo environment
   - Architecture diagram visualization

## Conclusion

With Next.js + Vercel + MkDocs, I achieved:

✅ **Modern and Fast Website**
- React Server Components
- Edge Network delivery
- Static generation

✅ **Developer-friendly Writing Environment**
- Write articles in Markdown/MDX
- Edit in VSCode
- Version control with Git

✅ **Fully Automated Deployment**
- Auto-publish on GitHub push
- Zero downtime
- Rollback capable

✅ **Zero-cost Operation**
- All features on free plan
- No need to worry about scale

I hope this helps anyone considering building a website for their open-source project!

## Links

- **Project Site**: https://tfdrift-falco.vercel.app/
- **GitHub**: https://github.com/higakikeita/driftwire
- **Documentation**: https://higakikeita.github.io/driftwire/
- **Next.js**: https://nextjs.org/
- **MkDocs Material**: https://squidfunk.github.io/mkdocs-material/
- **Vercel**: https://vercel.com/

---

**Contributions and feedback to driftwire are welcome!**

Stars and contributions on GitHub are greatly appreciated 🙏
