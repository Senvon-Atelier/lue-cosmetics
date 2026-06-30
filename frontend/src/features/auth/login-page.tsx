import { useState, FormEvent } from 'react';
import { useNavigate, Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Button } from '../shared/ui/button';

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
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
    <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="font-display text-4xl mb-2">Welcome Back</h1>
          <p className="text-ink-muted">Sign in to access your account</p>
        </div>

        {/* Error message */}
        {error && (
          <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Login form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Email field */}
          <div>
            <label htmlFor="email" className="block font-label font-medium text-ink mb-2">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="your@email.com"
            />
          </div>

          {/* Password field */}
          <div>
            <label htmlFor="password" className="block font-label font-medium text-ink mb-2">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="••••••••"
            />
          </div>

          {/* Forgot password link */}
          <div className="flex justify-end">
            <Link
              to="/forgot-password"
              className="text-sm text-lavender-600 hover:text-lavender-700 font-label transition-colors"
            >
              Forgot password?
            </Link>
          </div>

          {/* Submit button */}
          <Button
            type="submit"
            variant="primary"
            size="lg"
            isLoading={isLoading}
            className="w-full"
          >
            Sign In
          </Button>
        </form>

        {/* Sign up link */}
        <div className="text-center text-sm">
          <span className="text-ink-muted">Don't have an account? </span>
          <Link
            to="/signup"
            className="text-lavender-600 hover:text-lavender-700 font-label font-medium transition-colors"
          >
            Sign up
          </Link>
        </div>
      </div>
    </div>
  );
}
