/**
 * Why Falco? - The philosophy behind event-driven drift detection
 *
 * A narrative explaining why Falco is the ideal complement to Terraform
 * for real-time infrastructure change detection.
 */

interface WhyFalcoPageProps {
  onBack: () => void;
}

export default function WhyFalcoPage({ onBack }: WhyFalcoPageProps) {
  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-900 to-gray-800 text-gray-100">
      {/* Header */}
      <header className="sticky top-0 z-10 bg-gray-900/95 backdrop-blur border-b border-gray-700 px-6 py-4">
        <div className="max-w-3xl mx-auto flex items-center justify-between">
          <h1 className="text-xl font-bold text-white">Why Falco?</h1>
          <button
            onClick={onBack}
            className="px-4 py-2 text-sm bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors"
          >
            Back to Graph
          </button>
        </div>
      </header>

      {/* Content */}
      <main className="max-w-3xl mx-auto px-6 py-12 space-y-16">
        {/* Epigraph */}
        <section className="text-center space-y-4">
          <blockquote className="text-2xl font-light italic text-gray-300 leading-relaxed">
            "Terraform tells us <span className="text-blue-400 font-semibold not-italic">what should exist</span>.
            <br />
            Falco tells us <span className="text-red-400 font-semibold not-italic">what actually happened</span>."
          </blockquote>
        </section>

        {/* The Blueprint */}
        <section className="space-y-6">
          <h2 className="text-3xl font-bold text-white flex items-center gap-3">
            <span className="text-blue-400 text-4xl">&#x1f3d7;&#xfe0f;</span>
            The Perfect Blueprint
          </h2>
          <div className="space-y-4 text-lg text-gray-300 leading-relaxed">
            <p>
              Imagine a city with a brilliant architect who documents everything in a{' '}
              <strong className="text-blue-400">blueprint</strong> (Terraform).
              Every building, every road, every gate — perfectly mapped out.
            </p>
            <p>
              But one night, someone secretly replaces a gate. By morning, a lock has been
              added. Yet the blueprint shows no changes.
            </p>
            <p>
              The architect eventually notices: <em>"...it's different."</em> But it's too late.
              They can see <strong>what</strong> changed, but not <strong>who</strong> did it,{' '}
              <strong>when</strong>, or <strong>why</strong>.
            </p>
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-6 text-center">
              <p className="text-gray-400 text-base">The blueprint only speaks of results, not actions.</p>
            </div>
          </div>
        </section>

        {/* The Witness */}
        <section className="space-y-6">
          <h2 className="text-3xl font-bold text-white flex items-center gap-3">
            <span className="text-red-400 text-4xl">&#x1f441;&#xfe0f;</span>
            Enter Falco: The Witness
          </h2>
          <div className="space-y-4 text-lg text-gray-300 leading-relaxed">
            <p>
              So the city hires <strong className="text-red-400">Falco</strong> — not an
              architect, not a designer, but a <strong>witness</strong>.
            </p>
            <p>Falco's job is singular and essential:</p>
            <blockquote className="border-l-4 border-red-500 pl-6 py-2 text-xl italic text-gray-200">
              To observe the exact moment someone takes action.
            </blockquote>
            <div className="grid grid-cols-2 gap-4 mt-6">
              {[
                { icon: '\uD83D\uDC64', label: 'Who touched the gate' },
                { icon: '\u23F0', label: 'When they did it' },
                { icon: '\uD83D\uDEAA', label: 'Which gate it was' },
                { icon: '\uD83C\uDFAF', label: 'What their intent was' },
              ].map((item) => (
                <div key={item.label} className="bg-gray-800 border border-gray-700 rounded-lg p-4 flex items-center gap-3">
                  <span className="text-2xl">{item.icon}</span>
                  <span className="text-gray-200">{item.label}</span>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* The Meeting */}
        <section className="space-y-6">
          <h2 className="text-3xl font-bold text-white flex items-center gap-3">
            <span className="text-purple-400 text-4xl">&#x1f91d;</span>
            Blueprint Meets Witness
          </h2>
          <div className="space-y-4 text-lg text-gray-300 leading-relaxed">
            <p>
              The architect listens to Falco's report, opens the blueprint, and realizes:
            </p>
            <blockquote className="border-l-4 border-purple-500 pl-6 py-2 text-xl italic text-gray-200">
              "That change... doesn't exist in my blueprint."
            </blockquote>
            <div className="bg-gray-800 border border-gray-700 rounded-xl p-8 mt-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="text-center space-y-2">
                  <div className="text-4xl">&#x1f4d0;</div>
                  <h3 className="text-xl font-bold text-blue-400">The Blueprint</h3>
                  <p className="text-sm text-gray-400">Terraform</p>
                  <p className="text-gray-300">Knows <em>what should exist</em></p>
                </div>
                <div className="text-center space-y-2">
                  <div className="text-4xl">&#x1f441;&#xfe0f;</div>
                  <h3 className="text-xl font-bold text-red-400">The Witness</h3>
                  <p className="text-sm text-gray-400">Falco</p>
                  <p className="text-gray-300">Knows <em>what actually happened</em></p>
                </div>
              </div>
              <p className="text-center mt-6 text-gray-400">
                Neither alone can protect the city.
              </p>
            </div>
          </div>
        </section>

        {/* Comparison Table */}
        <section className="space-y-6">
          <h2 className="text-3xl font-bold text-white flex items-center gap-3">
            <span className="text-green-400 text-4xl">&#x26a1;</span>
            Traditional vs. Event-Driven
          </h2>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-gray-700">
                  <th className="py-3 px-4 text-gray-400 font-medium">Aspect</th>
                  <th className="py-3 px-4 text-gray-400 font-medium">Periodic Scan</th>
                  <th className="py-3 px-4 text-green-400 font-medium">TFDrift-Falco</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {[
                  ['Detection speed', 'Minutes to hours', 'Seconds'],
                  ['Who changed it?', 'Unknown', 'Full user identity'],
                  ['When?', 'Approximate', 'Exact timestamp'],
                  ['How?', 'Unknown', 'CloudTrail event detail'],
                  ['Mechanism', 'terraform plan / polling', 'Falco gRPC stream'],
                ].map(([aspect, traditional, tfdrift]) => (
                  <tr key={aspect} className="border-b border-gray-800">
                    <td className="py-3 px-4 font-medium text-gray-200">{aspect}</td>
                    <td className="py-3 px-4 text-gray-500">{traditional}</td>
                    <td className="py-3 px-4 text-green-300">{tfdrift}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        {/* Closing */}
        <section className="text-center space-y-6 pb-12">
          <div className="w-16 h-px bg-gray-700 mx-auto" />
          <p className="text-xl text-gray-300 leading-relaxed">
            Placing Falco between your infrastructure means
          </p>
          <p className="text-3xl font-bold text-white">
            Adding a <span className="text-red-400">witness</span> to your cloud.
          </p>
          <button
            onClick={onBack}
            className="mt-8 px-8 py-3 bg-red-600 hover:bg-red-500 text-white rounded-lg font-medium transition-colors"
          >
            See it in action
          </button>
        </section>
      </main>
    </div>
  );
}
