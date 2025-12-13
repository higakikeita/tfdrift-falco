import { redirect } from 'next/navigation'

export default function DocsPage() {
  // Redirect to the MkDocs documentation hosted on GitHub Pages
  redirect('https://higakikeita.github.io/tfdrift-falco/')
}
