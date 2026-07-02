// DORMANT: Tailwind is not wired into the build (no PostCSS/Vite plugin) and
// this config is currently unused. Kept for the future CSS→Tailwind migration.
// See docs/superpowers/specs/2026-07-02-funnel-ui-alignment-design.md §3.

import type { Config } from 'tailwindcss';

// Map legacy CSS custom properties to Tailwind theme
// Reference: Rue/styles.css lines 5-44
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      // Palette: --lavender-50 to --lavender-800, --cream, --paper, --ink, --ink-soft, --ink-muted, --line, --line-soft
      colors: {
        lavender: {
          50: 'var(--lavender-50)',
          100: 'var(--lavender-100)',
          200: 'var(--lavender-200)',
          300: 'var(--lavender-300)',
          400: 'var(--lavender-400)',
          500: 'var(--lavender-500)',
          600: 'var(--lavender-600)',
          700: 'var(--lavender-700)',
          800: 'var(--lavender-800)',
        },
        cream: 'var(--cream)',
        paper: 'var(--paper)',
        ink: 'var(--ink)',
        'ink-soft': 'var(--ink-soft)',
        'ink-muted': 'var(--ink-muted)',
        line: 'var(--line)',
        'line-soft': 'var(--line-soft)',
      },
      // Typography: --font-display, --font-serif, --font-body, --font-label
      fontFamily: {
        display: ['var(--font-display)', 'serif'],
        serif: ['var(--font-serif)', 'serif'],
        body: ['var(--font-body)', 'sans-serif'],
        label: ['var(--font-label)', 'sans-serif'],
      },
      // Motion: --ease: cubic-bezier(0.2, 0.8, 0.2, 1), --dur: 280ms
      easing: {
        default: 'var(--ease)',
      },
      duration: {
        DEFAULT: 'var(--dur)',
      },
      // Layout: --max: 1440px, --gut: clamp(20px, 4vw, 64px), --radius: 4px, --radius-lg: 14px
      maxWidth: {
        screen: 'var(--max)',
      },
      spacing: {
        gut: 'var(--gut)',
      },
      borderRadius: {
        DEFAULT: 'var(--radius)',
        lg: 'var(--radius-lg)',
      },
    },
  },
  plugins: [],
} satisfies Config;
