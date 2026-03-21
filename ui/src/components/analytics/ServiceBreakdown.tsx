import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface ServiceBreakdownProps {
  data: { service: string; count: number }[];
}

export function ServiceBreakdown({ data }: ServiceBreakdownProps) {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-5">
      <h3 className="text-sm font-semibold text-slate-700 mb-4">Top Affected Services</h3>
      <ResponsiveContainer width="100%" height={260}>
        <BarChart data={data} layout="vertical" margin={{ top: 5, right: 20, left: 10, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" horizontal={false} />
          <XAxis type="number" tick={{ fontSize: 12 }} stroke="#94a3b8" allowDecimals={false} />
          <YAxis type="category" dataKey="service" tick={{ fontSize: 11 }} stroke="#94a3b8" width={110} />
          <Tooltip contentStyle={{ borderRadius: 8, fontSize: 13, border: '1px solid #e2e8f0' }} />
          <Bar dataKey="count" fill="#6366f1" radius={[0, 4, 4, 0]} barSize={18} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
