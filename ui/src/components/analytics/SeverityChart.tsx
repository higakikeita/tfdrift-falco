import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';

interface SeverityChartProps {
  data: { name: string; value: number; fill: string }[];
}

export function SeverityChart({ data }: SeverityChartProps) {
  const total = data.reduce((s, d) => s + d.value, 0);

  return (
    <div className="bg-white rounded-xl border border-slate-200 p-5">
      <h3 className="text-sm font-semibold text-slate-700 mb-4">Severity Distribution</h3>
      <ResponsiveContainer width="100%" height={260}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={55}
            outerRadius={90}
            paddingAngle={3}
            dataKey="value"
            label={({ name, value }) => `${name} (${value})`}
            labelLine={{ stroke: '#94a3b8' }}
          >
            {data.map((entry, i) => (
              <Cell key={i} fill={entry.fill} />
            ))}
          </Pie>
          <Tooltip formatter={(value: number) => [`${value} events (${((value / total) * 100).toFixed(0)}%)`, '']} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
