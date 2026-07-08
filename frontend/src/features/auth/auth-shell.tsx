import type { ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { Icon } from '../shared/ui/icons';

interface AuthShellProps {
  title: ReactNode;
  sub: string;
  children: ReactNode;
  footer?: ReactNode;
}

export function AuthShell({ title, sub, children, footer }: AuthShellProps) {
  return (
    <div className="auth-wrap">
      <div className="auth-visual">
        <div className="ph ph--lavender"><span className="ph-label">editorial · lifestyle</span></div>
        <div className="auth-visual-copy">
          <div className="eyebrow" style={{ color: 'var(--lavender-300)' }}>Lue · Members</div>
          <h2>
            Small rituals,<br />
            <em style={{ fontFamily: 'var(--font-serif)', fontStyle: 'italic', color: 'var(--lavender-300)' }}>
              long kept.
            </em>
          </h2>
          <p>Track your orders, build your rituals, and earn points toward the shelf you've always wanted.</p>
        </div>
      </div>
      <div className="auth-form-wrap">
        <div className="auth-form">
          <Link className="back-link" to="/"><Icon name="arrowLeft" size={12} /> Back to Lue</Link>
          <div className="eyebrow" style={{ color: 'var(--lavender-700)' }}>{sub}</div>
          <h1>{title}</h1>
          {children}
          {footer}
        </div>
      </div>
    </div>
  );
}
