export function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900">Analytics</h1>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl border border-slate-200 p-6 h-80 flex items-center justify-center text-slate-400">
          Drift Trend Chart (Issue #22)
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-6 h-80 flex items-center justify-center text-slate-400">
          Provider Breakdown (Issue #22)
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-6 h-80 flex items-center justify-center text-slate-400">
          Service Breakdown (Issue #22)
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-6 h-80 flex items-center justify-center text-slate-400">
          Severity Distribution (Issue #22)
        </div>
      </div>
    </div>
  );
}
