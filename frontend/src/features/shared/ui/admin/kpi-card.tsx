interface KPICardProps {
  title: string;
  value: string | number;
}

export function KPICard({ title, value }: KPICardProps) {
  return (
    <div className="admin-kpi">
      <div className="admin-kpi-k">{title}</div>
      <div className="admin-kpi-v">{value}</div>
    </div>
  );
}
