export function EventsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Drift Events</h1>
        <div className="flex items-center gap-2">
          <select className="px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white">
            <option>All Providers</option>
            <option>AWS</option>
            <option>GCP</option>
          </select>
          <select className="px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white">
            <option>All Severities</option>
            <option>Critical</option>
            <option>High</option>
            <option>Medium</option>
            <option>Low</option>
          </select>
        </div>
      </div>

      <div className="bg-white rounded-xl border border-slate-200 p-6 min-h-[500px] flex items-center justify-center text-slate-400">
        Event Table with Filtering & Pagination (Issue #20)
      </div>
    </div>
  );
}
