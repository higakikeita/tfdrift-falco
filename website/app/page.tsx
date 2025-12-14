import Link from 'next/link'
import { GitHubIcon, RocketIcon, ShieldIcon, BoltIcon } from './components/Icons'

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900">
      {/* Navigation */}
      <nav className="border-b border-slate-700/50 backdrop-blur-sm bg-slate-900/80 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-2">
              <ShieldIcon className="w-8 h-8 text-indigo-400" />
              <span className="text-2xl font-bold text-white">TFDrift-Falco</span>
            </div>
            <div className="flex items-center space-x-6">
              <Link href="/blog" className="text-slate-300 hover:text-white transition-colors">
                Blog
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
                className="flex items-center space-x-2 text-slate-300 hover:text-white transition-colors"
              >
                <GitHubIcon className="w-5 h-5" />
                <span>GitHub</span>
              </a>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center space-y-8">
            <div className="inline-flex items-center space-x-2 px-4 py-2 bg-indigo-500/10 border border-indigo-500/20 rounded-full text-indigo-300 text-sm">
              <RocketIcon className="w-4 h-4" />
              <span>v0.3.0-dev - 164 CloudTrail Events Across 16 AWS Services</span>
            </div>

            <h1 className="text-5xl md:text-7xl font-bold text-white leading-tight">
              Real-time Terraform<br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-purple-400">
                Drift Detection
              </span>
            </h1>

            <p className="text-xl text-slate-300 max-w-3xl mx-auto leading-relaxed">
              Detect manual infrastructure changes instantly using Falco&apos;s CloudTrail plugin.
              No more periodic scans - get alerts the moment someone modifies your AWS resources.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <Link
                href="/docs"
                className="px-8 py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg font-semibold transition-all shadow-lg shadow-indigo-500/50 hover:shadow-xl hover:shadow-indigo-500/50"
              >
                Get Started
              </Link>
              <a
                href="https://github.com/higakikeita/tfdrift-falco"
                target="_blank"
                rel="noopener noreferrer"
                className="px-8 py-3 bg-slate-800 hover:bg-slate-700 text-white rounded-lg font-semibold transition-all border border-slate-600"
              >
                View on GitHub
              </a>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-slate-800/50">
        <div className="max-w-7xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold text-white text-center mb-16">
            Why TFDrift-Falco?
          </h2>

          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon={<BoltIcon className="w-8 h-8 text-indigo-400" />}
              title="Real-time Detection"
              description="Get instant alerts when infrastructure changes occur, powered by Falco's CloudTrail plugin."
            />
            <FeatureCard
              icon={<ShieldIcon className="w-8 h-8 text-indigo-400" />}
              title="Security Context"
              description="Track who made what changes with full IAM user identity and CloudTrail event correlation."
            />
            <FeatureCard
              icon={<RocketIcon className="w-8 h-8 text-indigo-400" />}
              title="164 CloudTrail Events"
              description="Monitor Lambda, EC2, ElastiCache, Auto Scaling, ECS, EKS, VPC, IAM, S3, and more across 16 AWS services."
            />
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="grid md:grid-cols-4 gap-8 text-center">
            <StatCard number="164" label="CloudTrail Events" />
            <StatCard number="16" label="AWS Services" />
            <StatCard number="100%" label="Test Coverage" />
            <StatCard number="83%" label="Roadmap Complete" />
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-r from-indigo-600 to-purple-600">
        <div className="max-w-4xl mx-auto text-center space-y-6">
          <h2 className="text-3xl md:text-4xl font-bold text-white">
            Ready to secure your infrastructure?
          </h2>
          <p className="text-xl text-indigo-100">
            Start monitoring your Terraform-managed resources in real-time.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Link
              href="/docs"
              className="px-8 py-3 bg-white text-indigo-600 rounded-lg font-semibold hover:bg-slate-100 transition-all shadow-lg"
            >
              Read Documentation
            </Link>
            <Link
              href="/releases"
              className="px-8 py-3 bg-indigo-800 text-white rounded-lg font-semibold hover:bg-indigo-900 transition-all"
            >
              View Releases
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-700/50 bg-slate-900 py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center text-slate-400">
            <p>Built with ❤️ by Keita Higaki</p>
            <p className="mt-2 text-sm">Licensed under MIT License</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="p-6 bg-slate-900/50 border border-slate-700/50 rounded-xl hover:border-indigo-500/50 transition-all">
      <div className="mb-4">{icon}</div>
      <h3 className="text-xl font-semibold text-white mb-2">{title}</h3>
      <p className="text-slate-400">{description}</p>
    </div>
  )
}

function StatCard({ number, label }: { number: string; label: string }) {
  return (
    <div className="space-y-2">
      <div className="text-4xl md:text-5xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-purple-400">
        {number}
      </div>
      <div className="text-slate-400">{label}</div>
    </div>
  )
}
