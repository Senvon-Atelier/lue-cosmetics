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
          <Icon name="user" size={14} /> Continue with Google
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
