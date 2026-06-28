// Icon component - minimal line icons, stroke-based
// Ported from Rue/src/icons.jsx

interface IconProps {
  name: keyof typeof iconPaths;
  size?: number;
  className?: string;
}

const iconPaths = {
  search: (
    <>
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </>
  ),
  heart: <path d="M12 21s-7-4.5-9.5-9A5.5 5.5 0 0 1 12 6a5.5 5.5 0 0 1 9.5 6C19 16.5 12 21 12 21z" />,
  bag: (
    <>
      <path d="M6 7h12l-1 13H7L6 7z" />
      <path d="M9 7V5a3 3 0 0 1 6 0v2" />
    </>
  ),
  menu: (
    <>
      <path d="M4 7h16" />
      <path d="M4 12h16" />
      <path d="M4 17h10" />
    </>
  ),
  close: (
    <>
      <path d="M6 6l12 12" />
      <path d="M18 6L6 18" />
    </>
  ),
  arrow: (
    <>
      <path d="M5 12h14" />
      <path d="m13 6 6 6-6 6" />
    </>
  ),
  arrowLeft: (
    <>
      <path d="M19 12H5" />
      <path d="m11 6-6 6 6 6" />
    </>
  ),
  arrowUp: <path d="m7 14 5-5 5 5" />,
  plus: (
    <>
      <path d="M12 5v14" />
      <path d="M5 12h14" />
    </>
  ),
  minus: <path d="M5 12h14" />,
  star: <path d="m12 3 2.6 5.9 6.4.6-4.8 4.3 1.4 6.3L12 17l-5.6 3.1 1.4-6.3L3 9.5l6.4-.6L12 3z" />,
  starFilled: <path d="m12 3 2.6 5.9 6.4.6-4.8 4.3 1.4 6.3L12 17l-5.6 3.1 1.4-6.3L3 9.5l6.4-.6L12 3z" fill="currentColor" />,
  truck: (
    <>
      <path d="M2 8h11v10H2zM13 11h5l3 3v4h-8z" />
      <circle cx="6" cy="19" r="2" />
      <circle cx="17" cy="19" r="2" />
    </>
  ),
  shield: <path d="M12 3 4 6v6c0 5 3.5 8 8 9 4.5-1 8-4 8-9V6l-8-3z" />,
  leaf: (
    <>
      <path d="M4 20c10 0 16-6 16-16-10 0-16 6-16 16z" />
      <path d="M4 20 20 4" />
    </>
  ),
  sparkle: (
    <>
      <path d="M12 3v18M3 12h18" />
      <path d="m6 6 12 12M6 18 18 6" strokeOpacity=".35" />
    </>
  ),
  filter: (
    <>
      <path d="M4 6h16" />
      <path d="M6 12h12" />
      <path d="M9 18h6" />
    </>
  ),
  grid: (
    <>
      <rect x="4" y="4" width="7" height="7" />
      <rect x="13" y="4" width="7" height="7" />
      <rect x="4" y="13" width="7" height="7" />
      <rect x="13" y="13" width="7" height="7" />
    </>
  ),
  list: (
    <>
      <path d="M4 6h16M4 12h16M4 18h16" />
    </>
  ),
  chevronDown: <path d="m6 9 6 6 6-6" />,
  chevronRight: <path d="m9 6 6 6-6 6" />,
  pin: (
    <>
      <path d="M12 21s-7-7-7-12a7 7 0 0 1 14 0c0 5-7 12-7 12z" />
      <circle cx="12" cy="9" r="2.5" />
    </>
  ),
  phone: <path d="M22 17v3a2 2 0 0 1-2 2A18 18 0 0 1 2 4a2 2 0 0 1 2-2h3a2 2 0 0 1 2 1.7l1 4a2 2 0 0 1-.5 2L8 11a16 16 0 0 0 5 5l1.3-1.5a2 2 0 0 1 2-.5l4 1a2 2 0 0 1 1.7 2z" />,
  clock: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </>
  ),
  check: <path d="m5 13 4 4L19 7" />,
  mail: (
    <>
      <rect x="3" y="5" width="18" height="14" />
      <path d="m3 7 9 6 9-6" />
    </>
  ),
  instagram: (
    <>
      <rect x="3" y="3" width="18" height="18" rx="5" />
      <circle cx="12" cy="12" r="4" />
      <circle cx="17.5" cy="6.5" r="0.6" fill="currentColor" />
    </>
  ),
  tiktok: <path d="M15 4v8.5a3.5 3.5 0 1 1-3.5-3.5M15 4c0 2.2 1.8 4 4 4" />,
  whatsapp: <path d="M20 12a8 8 0 0 1-11.8 7L4 20l1-4.2A8 8 0 1 1 20 12zM8.5 9c0 3.5 3 6.5 6.5 6.5l1.5-1-2-1-.8.7A6 6 0 0 1 10.3 11l.7-.8-1-2-1 1.5z" />,
  user: (
    <>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 21a8 8 0 0 1 16 0" />
    </>
  ),
  sliders: (
    <>
      <path d="M4 6h10M18 6h2M4 12h2M10 12h10M4 18h14M18 18h2" />
      <circle cx="16" cy="6" r="2" />
      <circle cx="8" cy="12" r="2" />
      <circle cx="16" cy="18" r="2" />
    </>
  ),
} as const;

export function Icon({ name, size = 20, className = '' }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      fill="none"
      stroke="currentColor"
      strokeWidth={1.4}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      {iconPaths[name]}
    </svg>
  );
}
