interface PanelProps {
  title?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
  /** Render children directly (mockup puts filter bars/tables flush inside .admin-panel) */
  flush?: boolean;
}

// Ported from Rue/admin.css .admin-panel structure
export function Panel({ title, actions, children, flush }: PanelProps) {
  return (
    <div className="admin-panel">
      {(title || actions) && (
        <div className="admin-panel-head">
          {title && <h3>{title}</h3>}
          {actions && <div className="admin-head-actions">{actions}</div>}
        </div>
      )}
      {flush ? children : <div className="admin-panel-body">{children}</div>}
    </div>
  );
}
