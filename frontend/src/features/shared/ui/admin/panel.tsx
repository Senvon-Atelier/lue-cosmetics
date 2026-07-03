interface PanelProps {
  title?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

// Ported from Rue/admin.css .admin-panel structure
export function Panel({ title, actions, children }: PanelProps) {
  return (
    <div className="admin-panel">
      {(title || actions) && (
        <div className="admin-panel-head">
          {title && <h3>{title}</h3>}
          {actions && <div className="admin-head-actions">{actions}</div>}
        </div>
      )}
      <div className="admin-panel-body">{children}</div>
    </div>
  );
}
