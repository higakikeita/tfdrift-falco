# TFDrift-Falco Website

Modern project website built with Next.js 14, TypeScript, and Tailwind CSS.

## 🚀 Features

- **Landing Page**: Beautiful hero section with feature showcase
- **Blog**: MDX-powered blog with syntax highlighting
- **Releases Page**: Automatic GitHub releases integration
- **Documentation**: Integrated with MkDocs on GitHub Pages
- **Google Analytics**: Built-in GA4 tracking support
- **Dark Mode**: Professional dark theme optimized for technical content
- **Responsive**: Mobile-first design
- **SEO Optimized**: Built-in Next.js SEO support with OpenGraph and Twitter cards

## 📦 Tech Stack

- Next.js 14 (App Router)
- TypeScript
- Tailwind CSS
- MDX (next-mdx-remote)
- Google Analytics 4 (@next/third-parties)
- Prism.js (Syntax highlighting)
- GitHub API Integration

## 🛠️ Development

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

Visit http://localhost:3000 (or 3001 if 3000 is in use)

## ⚙️ Configuration

### Environment Variables

Create a `.env.local` file for development:

```bash
# Google Analytics (optional)
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX

# Site URL (optional, for production)
NEXT_PUBLIC_SITE_URL=https://yourdomain.com
```

See `.env.example` for all available options.

### Blog Posts

Add blog posts in `content/blog/` as MDX files:

```mdx
---
title: "Your Post Title"
date: "2025-12-12"
excerpt: "Brief description"
author: "Your Name"
tags: ["tag1", "tag2"]
---

# Your Content Here

Write your blog post using Markdown and JSX...
```

## 🚢 Deployment

See detailed guides:

- **[Deployment Guide](./DEPLOYMENT.md)** - Complete deployment instructions
- **[Custom Domain Setup](./DOMAIN_SETUP.md)** - Configure your custom domain

### Quick Deploy to Vercel

```bash
npm i -g vercel
cd website
vercel --prod
```

## 📁 Project Structure

```
website/
├── app/
│   ├── page.tsx              # Landing page
│   ├── blog/                 # Blog
│   │   ├── page.tsx          # Blog listing
│   │   └── [slug]/           # Individual posts
│   │       └── page.tsx
│   ├── releases/             # Releases page
│   │   └── page.tsx
│   ├── docs/                 # Documentation redirect
│   │   └── page.tsx
│   ├── components/           # Reusable components
│   │   └── Icons.tsx
│   ├── layout.tsx            # Root layout with GA
│   └── globals.css           # Global styles + Prism theme
├── content/
│   └── blog/                 # Blog posts (MDX)
│       └── *.mdx
├── public/                   # Static assets
├── .env.example              # Environment variables template
├── DEPLOYMENT.md             # Deployment guide
├── DOMAIN_SETUP.md           # Custom domain guide
├── vercel.json               # Vercel configuration
└── package.json              # Dependencies
```

## 🔗 Links

- **GitHub**: https://github.com/higakikeita/tfdrift-falco
- **Documentation**: https://higakikeita.github.io/tfdrift-falco/
- **Docker**: https://ghcr.io/higakikeita/tfdrift-falco

## 📝 License

MIT License - see parent project for details
