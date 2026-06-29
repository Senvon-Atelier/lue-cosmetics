import { ReactNode } from 'react';

interface PageContainerProps {
  children: ReactNode;
  className?: string;
}

export function PageContainer({ children, className = '' }: PageContainerProps) {
  return (
    <div className={`wrap ${className}`} style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
      {children}
    </div>
  );
}
