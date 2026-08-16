import type { Metadata } from 'next'
import Link from 'next/link'
import { BoltIcon, BookIcon, GitHubIcon, LinkedInIcon, ShieldIcon, XIcon } from '../components/Icons'

export const metadata: Metadata = {
  title: 'Links — driftwire',
  description: 'Everything from the talk: code, documentation and contact for driftwire, Remedify and WhyQ.',
  // One QR code points here, so this page has to stay at this path.
  alternates: { canonical: 'https://tfdrift-falco.vercel.app/links' },
}

type LinkItem = {
  label: string
  href: string
  hint: string
  icon: (props: { className?: string }) => React.ReactElement
}

const sections: { heading: string; items: LinkItem[] }[] = [
  {
    heading: 'Code',
    items: [
      {
        label: 'driftwire',
        href: 'https://github.com/higakikeita/driftwire',
        hint: 'Real-time drift detection, on Falco',
        icon: ShieldIcon,
      },
      {
        label: 'Remedify',
        href: 'https://github.com/higakikeita/remedify',
        hint: 'Scan result → the exact fix command',
        icon: BoltIcon,
      },
      {
        label: 'WhyQ',
        href: 'https://github.com/higakikeita/whyq',
        hint: 'Any finding, in plain language',
        icon: BookIcon,
      },
    ],
  },
  {
    heading: 'Documentation',
    items: [
      {
        label: 'driftwire Docs',
        href: 'https://tfdrift-falco.vercel.app',
        hint: 'Getting started and architecture',
        icon: BookIcon,
      },
    ],
  },
  {
    heading: 'Connect',
    items: [
      { label: 'GitHub', href: 'https://github.com/higakikeita', hint: '@higakikeita', icon: GitHubIcon },
      { label: 'X', href: 'https://x.com/keitah0322', hint: '@keitah0322', icon: XIcon },
      {
        label: 'LinkedIn',
        href: 'https://www.linkedin.com/in/keita-higaki-a81377176',
        hint: 'Keita Higaki',
        icon: LinkedInIcon,
      },
    ],
  },
]

export default function LinksPage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900">
      <main className="mx-auto max-w-xl px-5 py-14 sm:py-20">
        <header className="space-y-3 text-center">
          <ShieldIcon className="mx-auto h-10 w-10 text-indigo-400" />
          <h1 className="text-2xl font-bold text-white sm:text-3xl">driftwire</h1>
          <p className="text-slate-300">Terraform Drift Detection with Falco</p>
        </header>

        {sections.map((section) => (
          <section key={section.heading} className="mt-10">
            <h2 className="mb-3 text-xs font-semibold tracking-widest text-slate-400 uppercase">
              {section.heading}
            </h2>
            <ul className="space-y-3">
              {section.items.map(({ label, href, hint, icon: Icon }) => (
                <li key={label}>
                  <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    // Tall rows so the whole card is an easy tap target on a phone.
                    className="flex items-center gap-4 rounded-xl border border-slate-700 bg-slate-800/60 px-5 py-4 transition-colors hover:border-indigo-500/60 hover:bg-slate-700/60"
                  >
                    <Icon className="h-6 w-6 shrink-0 text-indigo-400" />
                    <span className="min-w-0">
                      <span className="block font-semibold text-white">{label}</span>
                      <span className="block text-sm text-slate-400">{hint}</span>
                    </span>
                    <span aria-hidden className="ml-auto text-slate-500">
                      ↗
                    </span>
                  </a>
                </li>
              ))}
            </ul>
          </section>
        ))}

        <footer className="mt-12 space-y-3 border-t border-slate-700/60 pt-6 text-center">
          <Link href="/" className="text-sm text-slate-400 transition-colors hover:text-white">
            ← tfdrift-falco.vercel.app
          </Link>
          <p className="text-xs text-slate-500">
            Personal projects, MIT licensed. Views are my own — not a Sysdig product or position.
          </p>
        </footer>
      </main>
    </div>
  )
}
