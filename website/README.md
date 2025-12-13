# TFDrift-Falco Website

Modern project website built with Next.js 14, TypeScript, and Tailwind CSS.

## 🚀 Features

- **Landing Page**: Beautiful hero section with feature showcase
- **Releases Page**: Automatic GitHub releases integration
- **Dark Mode**: Professional dark theme optimized for technical content
- **Responsive**: Mobile-first design
- **SEO Optimized**: Built-in Next.js SEO support

## 📦 Tech Stack

- Next.js 14 (App Router)
- TypeScript
- Tailwind CSS
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

## 🚢 Deployment

### Deploy to Vercel (Recommended)

1. Install Vercel CLI:
```bash
npm i -g vercel
```

2. Deploy:
```bash
vercel
```

3. Follow the prompts to link your project

### Manual Deploy

1. Push to GitHub
2. Visit [vercel.com](https://vercel.com)
3. Import your repository
4. Vercel will auto-detect Next.js and deploy

## 📁 Project Structure

```
website/
├── app/
│   ├── page.tsx          # Landing page
│   ├── releases/         # Releases page
│   │   └── page.tsx
│   ├── components/       # Reusable components
│   │   └── Icons.tsx
│   └── layout.tsx        # Root layout
├── public/               # Static assets
├── vercel.json          # Vercel configuration
└── package.json
```

## 🔗 Links

- **GitHub**: https://github.com/higakikeita/tfdrift-falco
- **Documentation**: https://higakikeita.github.io/tfdrift-falco/
- **Docker**: https://ghcr.io/higakikeita/tfdrift-falco

## 📝 License

MIT License - see parent project for details
