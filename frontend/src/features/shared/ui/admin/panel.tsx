interface PanelProps {
  title?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export function Panel({ title, actions, children }: PanelProps) {
  return (
    <div className="bg-white border border-line rounded-xl mb-5 overflow-hidden">
      {(title || actions) && (
        <div className="flex items-center justify-between px-6 py-[18px] border-b border-line gap-3 flex-wrap">
          {title && <h3 className="font-display text-[17px] font-medium m-0">{title}</h3>}
          {actions && <div className="flex items-center gap-2">{actions}</div>}
        </div>
      )}
      <div className="p-5">{children}</div>
    </div>
  );
}
