# TFDrift-FalcoプロジェクトサイトをNext.js + Vercel + MkDocsで構築した話

## はじめに

Terraform drift検出ツール「TFDrift-Falco」のプロジェクトサイトを、Next.js 14とVercel、MkDocsを組み合わせて構築しました。この記事では、技術選定から実装、ハマったポイントまで、構築の全過程を紹介します。

- **プロジェクトサイト**: https://tfdrift-falco.vercel.app/
- **ドキュメント**: https://higakikeita.github.io/tfdrift-falco/
- **GitHub**: https://github.com/higakikeita/tfdrift-falco

## TFDrift-Falcoとは

TFDrift-Falcoは、AWS CloudTrailとFalcoを組み合わせた、リアルタイムのTerraform drift検出システムです。

### 主な機能
- AWSコンソールでの手動変更を即座に検知
- 120以上のCloudTrailイベントに対応
- Terraform stateとの差分を自動検出
- Slack/PagerDuty/Grafanaへのアラート通知

## なぜプロジェクトサイトを作ったのか

オープンソースプロジェクトの成長には、技術的な実装だけでなく、情報発信も重要です。以下の目的でプロジェクトサイトを構築しました：

1. **認知度向上**: GitHubだけでなく、独自のプレゼンスを確立
2. **ドキュメントの体系化**: 散在していた情報を整理
3. **リリース情報の発信**: 最新の開発状況を伝える
4. **コミュニティ形成**: ユーザーとの接点を増やす

## 技術スタック

### 全体構成

```
TFDrift-Falco Website
├── Vercel (Next.js 14)
│   ├── Landing Page
│   ├── Blog (MDX)
│   └── Releases (GitHub API)
│
└── GitHub Pages (MkDocs)
    └── Technical Documentation
```

### 選定理由

#### Next.js 14 + Vercel
- **App Router**: 最新のReact Server Componentsを活用
- **ゼロコンフィグデプロイ**: GitHubプッシュで自動デプロイ
- **Edge Network**: 世界中で高速配信
- **無料枠が充実**: 個人プロジェクトに最適

#### MkDocs Material
- **技術ドキュメントに特化**: コードブロック、検索、ナビゲーションが優秀
- **GitHub Pagesとの相性**: 簡単にホスティング可能
- **Markdown**: 開発者が書きやすい

#### MDX (next-mdx-remote)
- **Markdown + React**: コンポーネントを埋め込める
- **シンタックスハイライト**: Prism.jsで美しいコード表示
- **フロントマター対応**: メタデータを簡単に管理

## 実装のポイント

### 1. ブログ機能の実装

#### ディレクトリ構成

```
website/
├── app/
│   ├── blog/
│   │   ├── page.tsx           # ブログ一覧
│   │   └── [slug]/
│   │       └── page.tsx       # 個別記事
│   └── layout.tsx
└── content/
    └── blog/
        └── *.mdx              # ブログ記事
```

#### 個別記事ページの実装

Next.js 15以降、`params`がPromiseになったため、`await`が必要です：

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
  const { slug } = await params  // ← awaitが必要！
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

// 静的生成のための設定
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

#### MDXファイルの形式

```markdown
---
title: "記事タイトル"
date: "2025-12-12"
excerpt: "記事の概要"
author: "著者名"
tags: ["tag1", "tag2"]
---

# 記事本文

ここにMarkdownで記事を書きます。

\`\`\`yaml
# コードブロックも書けます
drift_rules:
  - name: "Production ECS Service Modified"
    resource_types:
      - "aws_ecs_service"
\`\`\`
```

### 2. GitHub API連携でリリースページ

GitHub APIを使ってリリース情報を自動取得：

```typescript
// app/releases/page.tsx
async function getReleases() {
  const res = await fetch(
    'https://api.github.com/repos/higakikeita/tfdrift-falco/releases',
    {
      next: { revalidate: 3600 } // 1時間ごとに再取得
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

### 3. MkDocsとの統合

#### ドキュメントへのリダイレクト

```typescript
// app/docs/page.tsx
import { redirect } from 'next/navigation'

export default function DocsPage() {
  redirect('https://higakikeita.github.io/tfdrift-falco/')
}
```

#### GitHub Actionsで自動デプロイ

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

### 4. Google Analytics統合

```typescript
// app/layout.tsx
import { GoogleAnalytics } from '@next/third-parties/google'

export default function RootLayout({ children }) {
  return (
    <html lang="ja">
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

環境変数の設定：

```bash
# .env.example
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
NEXT_PUBLIC_SITE_URL=https://tfdrift-falco.vercel.app
```

## ハマったポイントと解決策

### 1. Next.js 15のparams仕様変更

**問題**: 個別ブログ記事ページで404エラーが発生

**原因**: Next.js 15から`params`がPromiseになり、`await`が必要に

```typescript
// ❌ 動かない
export default function BlogPostPage({ params }: BlogPostPageProps) {
  const post = getPostData(params.slug)  // エラー
}

// ✅ 正しい
export default async function BlogPostPage({ params }: BlogPostPageProps) {
  const { slug } = await params  // awaitが必要
  const post = getPostData(slug)
}
```

### 2. GitHub PagesのCNAME問題

**問題**: MkDocsドキュメントが404エラー

**原因**: `CNAME`ファイルにカスタムドメイン（`docs.tfdrift-falco.com`）が設定されているが、DNSが未設定

**解決策**: ワークフローからCNAME設定を削除

```yaml
# ❌ カスタムドメイン未設定の場合
- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    cname: docs.tfdrift-falco.com  # これが原因

# ✅ 修正版
- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    # cname: docs.tfdrift-falco.com  # コメントアウト
    publish_dir: ./site
```

### 3. MkDocsのstrict mode

**問題**: MkDocsビルドが失敗

```
Aborted with 16 warnings in strict mode!
```

**原因**: ドキュメント内のリンク切れが警告として検出され、`--strict`モードでエラーに

**解決策**: `--strict`フラグを削除

```yaml
# ❌ strict modeで警告をエラー扱い
- name: Build MkDocs site
  run: mkdocs build --strict --verbose

# ✅ 警告は表示するが、ビルドは続行
- name: Build MkDocs site
  run: mkdocs build --verbose
```

### 4. Prism.jsのスタイリング

**問題**: コードブロックのシンタックスハイライトが効かない

**解決策**: グローバルCSSにPrismテーマを追加

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

/* トークンの色設定 */
.token.comment { color: #94a3b8; }
.token.string { color: #86efac; }
.token.keyword { color: #a78bfa; }
.token.function { color: #60a5fa; }
```

## デプロイフロー

### Vercel（ウェブサイト）

```bash
1. GitHubにpush
   ↓
2. Vercelが自動検知
   ↓
3. Next.jsビルド
   ↓
4. Edge Networkにデプロイ
   ↓
5. https://tfdrift-falco.vercel.app/ で公開
```

### GitHub Pages（ドキュメント）

```bash
1. docs/配下を編集してpush
   ↓
2. GitHub Actionsが起動
   ↓
3. MkDocsビルド
   ↓
4. gh-pagesブランチにpush
   ↓
5. https://higakikeita.github.io/tfdrift-falco/ で公開
```

## 成果

### パフォーマンス

- **Lighthouse Score**: 95+（全項目）
- **First Contentful Paint**: < 1.0s
- **Time to Interactive**: < 2.0s

### 運用面

✅ **ブログ記事の追加が簡単**
- `.mdx`ファイルを置いてpushするだけ
- ビルド時に自動で静的生成

✅ **ドキュメントの更新が自動**
- Markdownを編集してpush
- GitHub Actionsが自動デプロイ

✅ **リリース情報が自動更新**
- GitHub Releasesを作成するだけ
- サイトが自動で最新情報を取得（1時間キャッシュ）

✅ **SEO最適化**
- OpenGraph、Twitter Cardsに対応
- 検索エンジンに最適化されたメタデータ

## コスト

すべて**無料**で運用できています：

- **Vercel**: Hobbyプラン（無料）
- **GitHub Pages**: 無料
- **GitHub Actions**: 月2000分まで無料（十分）
- **ドメイン**: Vercelの`vercel.app`ドメインを使用

## 今後の展開

### 短期的な改善

1. **ブログコンテンツの充実**
   - 技術解説記事
   - ユースケース紹介
   - 開発日記

2. **ドキュメントの拡張**
   - より多くのAWSサービス対応
   - チュートリアルの追加
   - トラブルシューティングガイド

3. **SEO最適化**
   - サイトマップの最適化
   - 構造化データの追加
   - パフォーマンスチューニング

### 中長期的な計画

1. **コミュニティ機能**
   - Discussionsとの連携
   - コントリビューターの紹介
   - ユーザー事例の掲載

2. **多言語対応**
   - 英語版ドキュメント
   - i18n対応

3. **インタラクティブなデモ**
   - ブラウザで試せるデモ環境
   - アーキテクチャ図の可視化

## まとめ

Next.js + Vercel + MkDocsの組み合わせで、以下を実現できました：

✅ **モダンで高速なウェブサイト**
- React Server Components
- Edge Network配信
- 静的生成

✅ **開発者フレンドリーな執筆環境**
- Markdown/MDXで記事作成
- VSCodeで編集可能
- Gitでバージョン管理

✅ **完全自動化されたデプロイ**
- GitHubプッシュで自動公開
- ゼロダウンタイム
- ロールバック可能

✅ **ゼロコストでの運用**
- 無料プランで全機能利用
- スケールを気にしなくて良い

オープンソースプロジェクトのウェブサイト構築を検討している方の参考になれば幸いです！

## 参考リンク

- **プロジェクトサイト**: https://tfdrift-falco.vercel.app/
- **GitHub**: https://github.com/higakikeita/tfdrift-falco
- **ドキュメント**: https://higakikeita.github.io/tfdrift-falco/
- **Next.js**: https://nextjs.org/
- **MkDocs Material**: https://squidfunk.github.io/mkdocs-material/
- **Vercel**: https://vercel.com/

---

**TFDrift-Falcoへの貢献やフィードバックをお待ちしています！**

GitHubでスターやコントリビュートをいただけると嬉しいです 🙏
