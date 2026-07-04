import { useState, FormEvent } from 'react';
import { useNavigate, Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Icon } from '../shared/ui/icons';
import { AuthShell } from './auth-shell';

const GOOGLE_START_URL = `${import.meta.env.VITE_API_URL ?? '/api/v1'}/auth/google/start`;

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [show, setShow] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      await login(email, password);
      // Redirect to account page or home on successful login
      navigate({ to: '/account' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid email or password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell
      title="Welcome back."
      sub="Sign in"
      footer={
        <div className="auth-meta">
          New here? <Link to="/signup">Create an account</Link>
        </div>
      }
    >
      <div className="auth-social">
        <button type="button" onClick={() => { window.location.href = GOOGLE_START_URL; }}>
          <svg width="14" height="14" viewBox="0 0 18 18" aria-hidden="true">
            <path fill="#4285F4" d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62z" />
            <path fill="#34A853" d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.84.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18z" />
            <path fill="#FBBC05" d="M3.97 10.72a5.4 5.4 0 0 1 0-3.44V4.95H.96a9 9 0 0 0 0 8.1l3.01-2.33z" />
            <path fill="#EA4335" d="M9 3.58c1.32 0 2.5.45 3.44 1.35l2.58-2.59A9 9 0 0 0 .96 4.95l3.01 2.33C4.68 5.16 6.66 3.58 9 3.58z" />
          </svg> Continue with Google
        </button>
      </div>
      <div className="auth-divider">or with email</div>
      <form onSubmit={handleSubmit}>
        <div className="field" style={{ marginBottom: 16 }}>
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            placeholder="you@somewhere.com"
          />
        </div>
        <div className="field" style={{ marginBottom: 8 }}>
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type={show ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            placeholder="••••••••"
          />
        </div>
        {error && (
          <div className="auth-meta" style={{ color: 'var(--lavender-700)', marginBottom: 8 }}>
            {error}
          </div>
        )}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
          <label style={{ display: 'inline-flex', gap: 8, fontFamily: 'var(--font-label)', fontSize: 12, color: 'var(--ink-soft)' }}>
            <input type="checkbox" onChange={(e) => setShow(e.target.checked)} /> Show password
          </label>
          <Link className="auth-link" to="/forgot-password">Forgot password?</Link>
        </div>
        <button
          type="submit"
          className="btn btn-primary"
          style={{ width: '100%', justifyContent: 'center' }}
          disabled={isLoading}
        >
          Sign in <Icon name="arrow" size={14} />
        </button>
      </form>
    </AuthShell>
  );
}
