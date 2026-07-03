import { Link } from '@tanstack/react-router';
import { useAuth } from '../../lib/auth/auth-provider';
import { AcctHead } from './acct-primitives';

export function AccountSettings() {
  const { user } = useAuth();

  return (
    <main className="acct-main">
      <AcctHead eyebrow="Settings" title="Profile" />

      <div className="form-card" style={{ marginBottom: 24 }}>
        <div className="alert alert-info">
          Profile editing isn't available yet.
        </div>
        <div className="form-row">
          <div className="field">
            <label>Name</label>
            <input value={user?.name || ''} readOnly placeholder="Not set" />
          </div>
          <div className="field">
            <label>Email</label>
            <input type="email" value={user?.email || ''} readOnly />
            <span className="field-hint">
              {user?.email_verified
                ? 'Verified'
                : 'Unverified — check your inbox for the verification email'}
              {' · '}changes require verification
            </span>
          </div>
        </div>
      </div>

      <div className="form-card" style={{ marginBottom: 24 }}>
        <h3 className="form-card-title">Password</h3>
        <div className="alert alert-info">
          Password changes aren't available from settings yet — use the email
          reset flow instead.
        </div>
        <Link className="btn btn-ghost" to="/forgot-password">
          Reset via email
        </Link>
      </div>

      <div className="form-card">
        <h3 className="form-card-title">Delete account</h3>
        <div className="alert alert-warn">
          Account deletion isn't available yet. Contact us if you need your
          data removed.
        </div>
      </div>
    </main>
  );
}
