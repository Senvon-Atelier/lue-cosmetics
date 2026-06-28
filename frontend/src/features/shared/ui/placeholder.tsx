// Placeholder imagery system
// Ported from legacy .ph CSS classes - used when product images are not available
// The tone-based colored rectangles with striped gradients and labels

interface PlaceholderProps {
  width?: number | string;
  height?: number | string;
  tone?: 'lavender' | 'cream' | 'ink' | 'rose';
  label?: string;
  className?: string;
  style?: React.CSSProperties;
}

const toneColors = {
  lavender: {
    bg: 'var(--lavender-100)',
    stripe1: 'var(--lavender-50)',
    stripe2: 'var(--lavender-200)',
  },
  cream: {
    bg: 'var(--cream)',
    stripe1: '#F5F2EB',
    stripe2: '#E8E5DD',
  },
  ink: {
    bg: 'var(--ink-soft)',
    stripe1: 'var(--ink-muted)',
    stripe2: '#2D2539',
  },
  rose: {
    bg: '#FCE7F1',
    stripe1: '#FFF0F5',
    stripe2: '#F9D5E7',
  },
};

export function ProductPlaceholder({
  width = '100%',
  height = '100%',
  tone = 'lavender',
  label = '',
  className = '',
  style = {},
}: PlaceholderProps) {
  const colors = toneColors[tone];

  return (
    <div
      className={`ph ph--${tone} ${className}`}
      style={{
        width: typeof width === 'number' ? `${width}px` : width,
        height: typeof height === 'number' ? `${height}px` : height,
        backgroundColor: colors.bg,
        backgroundImage: `repeating-linear-gradient(
          45deg,
          ${colors.stripe1} 0px,
          ${colors.stripe1} 10px,
          ${colors.stripe2} 10px,
          ${colors.stripe2} 20px
        )`,
        backgroundSize: '20px 20px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
        ...style,
      }}
    >
      {label && (
        <span
          className="ph-label"
          style={{
            fontSize: '0.75rem',
            fontWeight: 500,
            color: 'var(--ink)',
            opacity: 0.6,
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            padding: '0.5rem 1rem',
            backgroundColor: 'rgba(255, 255, 255, 0.8)',
            borderRadius: 'var(--radius)',
            backdropFilter: 'blur(4px)',
          }}
        >
          {label}
        </span>
      )}
    </div>
  );
}
