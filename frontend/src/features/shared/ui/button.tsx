import { Icon } from './icons';

// Button component
// Ported from legacy .btn CSS classes

type IconName = 'search' | 'heart' | 'bag' | 'menu' | 'close' | 'arrow' | 'arrowLeft' | 'arrowUp' | 'plus' | 'minus' | 'star' | 'starFilled' | 'truck' | 'shield' | 'leaf' | 'sparkle' | 'filter' | 'grid' | 'list' | 'chevronDown' | 'chevronRight' | 'pin' | 'phone' | 'clock' | 'check' | 'mail' | 'instagram' | 'tiktok' | 'whatsapp' | 'user' | 'sliders';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  icon?: IconName | 'none';
  iconPosition?: 'left' | 'right';
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  isLoading = false,
  icon = 'none',
  iconPosition = 'right',
  className = '',
  disabled,
  ...props
}: ButtonProps) {
  const baseStyles = 'btn inline-flex items-center justify-center gap-2 font-label font-medium transition-all duration-[var(--dur)] ease-[var(--ease)] focus:outline-none focus:ring-2 focus:ring-lavender-400 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed';

  const variantStyles = {
    primary: 'bg-ink text-paper hover:bg-ink-soft active:bg-ink',
    secondary: 'bg-lavender-200 text-ink hover:bg-lavender-300 active:bg-lavender-400',
    outline: 'border border-line text-ink hover:border-lavender-400 hover:bg-lavender-50 active:bg-lavender-100',
    ghost: 'text-ink hover:bg-lavender-100 active:bg-lavender-200',
  };

  const sizeStyles = {
    sm: 'text-sm px-3 py-1.5 rounded',
    md: 'text-base px-4 py-2 rounded-lg',
    lg: 'text-lg px-6 py-3 rounded-lg',
  };

  return (
    <button
      className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? (
        <>
          <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
          Loading...
        </>
      ) : (
        <>
          {icon !== 'none' && iconPosition === 'left' && <Icon name={icon} size={16} />}
          {children}
          {icon !== 'none' && iconPosition === 'right' && <Icon name={icon} size={16} />}
        </>
      )}
    </button>
  );
}
