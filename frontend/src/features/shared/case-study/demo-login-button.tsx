import { Icon } from '../ui/icons';

interface DemoLoginButtonProps {
  onLogin: (email: string, password: string) => void;
}

export function DemoLoginButton({ onLogin }: DemoLoginButtonProps) {
  return (
    <button
      type="button"
      className="btn btn-secondary cs-demo-btn"
      onClick={() => onLogin('customer@ruecosmetics.com', 'hunter22')}
    >
      <Icon name="sparkle" size={14} /> Demo Login
    </button>
  );
}
