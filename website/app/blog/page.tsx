import Link from 'next/link'
import fs from 'fs'
import path from 'path'
import matter from 'gray-matter'

interface BlogPost {
  slug: string
  title: string
  date: string
  excerpt: string
  author: string
  tags: string[]
}

function getBlogPosts(): BlogPost[] {
  const postsDirectory = path.join(process.cwd(), 'content/blog')
  
  if (!fs.existsSync(postsDirectory)) {
    return []
  }
  
  const filenames = fs.readdirSync(postsDirectory)
  
  const posts = filenames
    .filter(filename => filename.endsWith('.mdx'))
    .map(filename => {
      const filePath = path.join(postsDirectory, filename)
      const fileContents = fs.readFileSync(filePath, 'utf8')
      const { data } = matter(fileContents)
      
      return {
        slug: filename.replace('.mdx', ''),
        title: data.title || 'Untitled',
        date: data.date || new Date().toISOString(),
        excerpt: data.excerpt || '',
        author: data.author || 'TFDrift-Falco Team',
        tags: data.tags || []
      }
    })
    .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
  
  return posts
}

export default function BlogPage() {
  const posts = getBlogPosts()
  
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900">
      {/* Navigation */}
      <nav className="border-b border-slate-700/50 backdrop-blur-sm bg-slate-900/80">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link href="/" className="text-2xl font-bold text-white hover:text-indigo-400 transition-colors">
              TFDrift-Falco
            </Link>
            <div className="flex items-center space-x-6">
              <Link href="/" className="text-slate-300 hover:text-white transition-colors">
                Home
              </Link>
              <Link href="/releases" className="text-slate-300 hover:text-white transition-colors">
                Releases
              </Link>
              <Link href="/docs" className="text-slate-300 hover:text-white transition-colors">
                Docs
              </Link>
              <a
                href="https://github.com/higakikeita/tfdrift-falco"
                target="_blank"
                rel="noopener noreferrer"
                className="text-slate-300 hover:text-white transition-colors"
              >
                GitHub
              </a>
            </div>
          </div>
        </div>
      </nav>

      {/* Header */}
      <div className="py-16 px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Blog
          </h1>
          <p className="text-xl text-slate-300">
            Latest updates, tutorials, and insights about TFDrift-Falco
          </p>
        </div>
      </div>

      {/* Blog Posts */}
      <div className="pb-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto">
          {posts.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-slate-400 mb-4">No blog posts yet. Check back soon!</p>
              <p className="text-sm text-slate-500">Posts will appear here once published.</p>
            </div>
          ) : (
            <div className="space-y-8">
              {posts.map((post) => (
                <BlogPostCard key={post.slug} post={post} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function BlogPostCard({ post }: { post: BlogPost }) {
  const date = new Date(post.date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })

  return (
    <Link href={`/blog/${post.slug}`}>
      <article className="bg-slate-900/50 border border-slate-700/50 rounded-xl p-6 hover:border-indigo-500/50 transition-all cursor-pointer">
        <div className="flex items-center space-x-4 text-sm text-slate-400 mb-3">
          <time dateTime={post.date}>{date}</time>
          <span>•</span>
          <span>{post.author}</span>
        </div>
        
        <h2 className="text-2xl font-bold text-white mb-3 hover:text-indigo-400 transition-colors">
          {post.title}
        </h2>
        
        <p className="text-slate-300 mb-4">
          {post.excerpt}
        </p>
        
        {post.tags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {post.tags.map((tag) => (
              <span
                key={tag}
                className="px-3 py-1 bg-indigo-500/10 border border-indigo-500/20 rounded-full text-indigo-300 text-sm"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </article>
    </Link>
  )
}
