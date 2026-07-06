// Logo (original pinwheel-circle mark, not the brand's exact glyph)
// Ported from Rue/src/shared.jsx

interface RueMarkProps {
  size?: number;
  color?: string;
}

export function RueMark({ size = 32, color = 'currentColor' }: RueMarkProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 40 40" fill="none" style={{ color }}>
      <circle cx="20" cy="20" r="18" stroke="currentColor" strokeWidth="1" opacity={0.4} />
      <circle cx="20" cy="20" r="13.5" stroke="currentColor" strokeWidth="1.2" />
      {/* abstract 4-blade pinwheel */}
      <g stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" fill="none">
        <path d="M20 20 C 20 14, 14 12, 12 14 C 14 18, 18 20, 20 20 Z" />
        <path d="M20 20 C 26 20, 28 14, 26 12 C 22 14, 20 18, 20 20 Z" />
        <path d="M20 20 C 20 26, 26 28, 28 26 C 26 22, 22 20, 20 20 Z" />
        <path d="M20 20 C 14 20, 12 26, 14 28 C 18 26, 20 22, 20 20 Z" />
      </g>
      <circle cx="20" cy="20" r="1.6" fill="currentColor" />
    </svg>
  );
}

export function Brand() {
  return (
    <>
      <div className="brand-mark">
        <RueMark size={22} />
      </div>
      <div>
        <div className="brand-word">Rue</div>
        <div className="brand-tag">Cosmetics</div>
      </div>
    </>
  );
}
