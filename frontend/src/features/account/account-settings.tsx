import { useAuth } from '../../lib/auth/auth-provider';
import { Button } from '../shared/ui/button';

export function AccountSettings() {
  const { user } = useAuth();
  const name = user?.name || '';
  const email = user?.email || '';

  return (
    <div>
      <div className="mb-8">
        <h2 className="font-display text-xl mb-2">Account Settings</h2>
        <p className="text-ink-muted">Manage your profile and security.</p>
      </div>

      {/* Profile Settings */}
      <div className="bg-white rounded-lg p-6 mb-6" style={{ border: '1px solid var(--line)' }}>
        <h3 className="font-label font-semibold mb-4">Profile Information</h3>
        <div className="mb-4 px-4 py-3 rounded-lg bg-lavender-50 border border-lavender-100 text-ink-soft text-sm">
          Profile editing is not available yet.
        </div>

        <div className="space-y-4">
          {/* Name */}
          <div>
            <label className="block font-label font-medium text-ink mb-2">
              Name
            </label>
            <input
              type="text"
              value={name}
              readOnly
              className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink-soft cursor-not-allowed"
              placeholder="Your full name"
            />
          </div>

          {/* Email */}
          <div>
            <label className="block font-label font-medium text-ink mb-2">
              Email
            </label>
            <input
              type="email"
              value={email}
              disabled
              className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink-soft cursor-not-allowed"
              title="Email changes require verification"
            />
            <p className="text-xs text-ink-muted mt-1">
              Email changes require verification flow
            </p>
          </div>

          <Button
            type="button"
            variant="primary"
            disabled
          >
            Save Profile Unavailable
          </Button>
        </div>
      </div>

      {/* Password Change */}
      <div className="bg-white rounded-lg p-6" style={{ border: '1px solid var(--line)' }}>
        <h3 className="font-label font-semibold mb-4">Change Password</h3>
        <div className="mb-4 px-4 py-3 rounded-lg bg-lavender-50 border border-lavender-100 text-ink-soft text-sm">
          Password changes are not available from account settings yet. Use the password reset flow from the login page.
        </div>

        <div className="space-y-4">
          {/* Current Password */}
          <div>
            <label className="block font-label font-medium text-ink mb-2">
              Current Password
            </label>
            <input
              type="password"
              disabled
              className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink-soft cursor-not-allowed"
              placeholder="Enter current password"
            />
          </div>

          {/* New Password */}
          <div>
            <label className="block font-label font-medium text-ink mb-2">
              New Password
            </label>
            <input
              type="password"
              disabled
              className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink-soft cursor-not-allowed"
              placeholder="Enter new password (min 8 characters)"
            />
          </div>

          {/* Confirm New Password */}
          <div>
            <label className="block font-label font-medium text-ink mb-2">
              Confirm New Password
            </label>
            <input
              type="password"
              disabled
              className="w-full px-4 py-2 border border-line rounded-lg bg-paper text-ink-soft cursor-not-allowed"
              placeholder="Confirm new password"
            />
          </div>

          <Button
            type="submit"
            variant="primary"
            disabled
          >
            Password Change Unavailable
          </Button>
        </div>
      </div>

      {/* Account Deletion */}
      <div className="bg-rose-50 border border-rose-200 rounded-lg p-6">
        <h3 className="font-label font-semibold mb-2 text-rose-800">Danger Zone</h3>
        <p className="text-rose-700 text-sm mb-4">
          Once you delete your account, there is no going back. Please be certain.
        </p>
        <Button
          variant="outline"
          size="sm"
          className="border-rose-300 text-rose-700 hover:bg-rose-50"
          disabled
        >
          Account Deletion Unavailable
        </Button>
      </div>
    </div>
  );
}
