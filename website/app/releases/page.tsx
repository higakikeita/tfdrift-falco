import Link from 'next/link'

interface Release {
  id: number
  tag_name: string
  name: string
  published_at: string
  body: string
  html_url: string
  prerelease: boolean
}

async function getReleases(): Promise<Release[]> {
  const res = await fetch('https://api.github.com/repos/higakikeita/tfdrift-falco/releases', {
    next: { revalidate: 3600 } // Revalidate every hour
  })
  
  if (!res.ok) {
    return []
  }
  
  return res.json()
}

export default async function ReleasesPage() {
  const releases = await getReleases()
  
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
            Releases
          </h1>
          <p className="text-xl text-slate-300">
            Track the evolution of TFDrift-Falco with our release history
          </p>
        </div>
      </div>

      {/* Releases List */}
      <div className="pb-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto space-y-6">
          {releases.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-slate-400">No releases found</p>
            </div>
          ) : (
            releases.map((release) => (
              <ReleaseCard key={release.id} release={release} />
            ))
          )}
        </div>
      </div>
    </div>
  )
}

function ReleaseCard({ release }: { release: Release }) {
  const date = new Date(release.published_at).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })

  return (
    <div className="bg-slate-900/50 border border-slate-700/50 rounded-xl p-6 hover:border-indigo-500/50 transition-all">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-4">
        <div className="flex items-center space-x-3">
          <h2 className="text-2xl font-bold text-white">
            {release.name || release.tag_name}
          </h2>
          {release.prerelease && (
            <span className="px-3 py-1 bg-yellow-500/10 border border-yellow-500/20 rounded-full text-yellow-300 text-sm">
              Pre-release
            </span>
          )}
        </div>
        <div className="text-slate-400 text-sm mt-2 md:mt-0">{date}</div>
      </div>

      <div className="prose prose-invert max-w-none mb-4">
        <div 
          className="text-slate-300 text-sm leading-relaxed"
          dangerouslySetInnerHTML={{ 
            __html: release.body
              ? release.body
                  .replace(/^### /gm, '<h3 class="text-lg font-semibold text-white mt-4 mb-2">')
                  .replace(/^## /gm, '<h2 class="text-xl font-semibold text-white mt-6 mb-3">')
                  .replace(/^# /gm, '<h1 class="text-2xl font-bold text-white mt-8 mb-4">')
                  .replace(/\n/g, '<br />')
              : 'No description provided'
          }}
        />
      </div>

      <div className="flex items-center space-x-4">
        <a
          href={release.html_url}
          target="_blank"
          rel="noopener noreferrer"
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition-all"
        >
          View on GitHub
        </a>
        <a
          href={`https://github.com/higakikeita/tfdrift-falco/releases/download/${release.tag_name}/tfdrift-linux-amd64`}
          className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg text-sm font-semibold transition-all border border-slate-600"
        >
          Download
        </a>
      </div>
    </div>
  )
}
