import { useState, FormEvent } from 'react';
import { useNavigate, Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { Button } from '../shared/ui/button';

export function SignupPage() {
  const navigate = useNavigate();
  const { signup } = useAuth();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [agreeToTerms, setAgreeToTerms] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Password strength checker
  const getPasswordStrength = (pwd: string): { strength: string; score: number } => {
    let score = 0;
    if (pwd.length >= 8) score++;
    if (pwd.length >= 12) score++;
    if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++;
    if (/\d/.test(pwd)) score++;
    if (/[!@#$%^&*(),.?":{}|<>]/.test(pwd)) score++;

    if (score <= 2) return { strength: 'Weak', score: 1 };
    if (score === 3) return { strength: 'Fair', score: 2 };
    if (score === 4) return { strength: 'Good', score: 3 };
    return { strength: 'Strong', score: 4 };
  };

  const passwordStrength = getPasswordStrength(password);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    if (!agreeToTerms) {
      setError('Please agree to the Terms of Service');
      return;
    }

    setIsLoading(true);

    try {
      await signup(email, password, name);
      // Redirect to account page or show verification message
      navigate({ to: '/account' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create account');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-paper text-ink font-body flex items-center justify-center py-12 px-4">
      <div className="max-w-md w-full space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="font-display text-4xl mb-2">Create Account</h1>
          <p className="text-ink-muted">Join us for exclusive offers and seamless shopping</p>
        </div>

        {/* Error message */}
        {error && (
          <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Signup form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Name field */}
          <div>
            <label htmlFor="name" className="block font-label font-medium text-ink mb-2">
              Full Name
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="Jane Doe"
            />
          </div>

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

            {/* Password strength indicator */}
            {password && (
              <div className="mt-2">
                <div className="flex gap-1 mb-1">
                  {[1, 2, 3, 4].map((level) => (
                    <div
                      key={level}
                      className={`h-1 flex-1 rounded-full transition-colors ${
                        passwordStrength.score >= level
                          ? level === 1
                            ? 'bg-rose-400'
                            : level === 2
                            ? 'bg-yellow-400'
                            : level === 3
                            ? 'bg-lavender-400'
                            : 'bg-green-400'
                          : 'bg-gray-200'
                      }`}
                    />
                  ))}
                </div>
                <p className="text-xs text-ink-muted">
                  Password strength: <span className="font-medium">{passwordStrength.strength}</span>
                </p>
              </div>
            )}
          </div>

          {/* Confirm password field */}
          <div>
            <label htmlFor="confirmPassword" className="block font-label font-medium text-ink mb-2">
              Confirm Password
            </label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              minLength={8}
              className="w-full px-4 py-2 border-b border-line bg-transparent text-ink placeholder:text-ink-muted focus:outline-none focus:border-lavender-400 focus:ring-1 focus:ring-lavender-400 transition-colors"
              placeholder="••••••••"
            />
            {confirmPassword && password !== confirmPassword && (
              <p className="mt-1 text-xs text-rose-600">Passwords do not match</p>
            )}
          </div>

          {/* Terms agreement */}
          <div className="flex items-start gap-3">
            <input
              id="terms"
              type="checkbox"
              checked={agreeToTerms}
              onChange={(e) => setAgreeToTerms(e.target.checked)}
              required
              className="mt-1 w-4 h-4 text-lavender-600 border-gray-300 rounded focus:ring-lavender-400 focus:ring-2"
            />
            <label htmlFor="terms" className="text-sm text-ink-soft">
              I agree to the{' '}
              <a href="#terms" className="text-lavender-600 hover:text-lavender-700">
                Terms of Service
              </a>{' '}
              and{' '}
              <a href="#privacy" className="text-lavender-600 hover:text-lavender-700">
                Privacy Policy
              </a>
            </label>
          </div>

          {/* Submit button */}
          <Button
            type="submit"
            variant="primary"
            size="lg"
            isLoading={isLoading}
            disabled={!agreeToTerms || password !== confirmPassword || password.length < 8}
            className="w-full"
          >
            Create Account
          </Button>
        </form>

        {/* Login link */}
        <div className="text-center text-sm">
          <span className="text-ink-muted">Already have an account? </span>
          <Link
            to="/login"
            className="text-lavender-600 hover:text-lavender-700 font-label font-medium transition-colors"
          >
            Log in
          </Link>
        </div>
      </div>
    </div>
  );
}
