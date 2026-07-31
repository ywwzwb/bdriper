/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        bg: '#0F172A',
        fg: '#F8FAFC',
        card: '#1B2336',
        muted: '#272F42',
        border: '#475569',
        accent: '#22C55E',
        destructive: '#EF4444',
        primary: '#1E293B',
        secondary: '#334155',
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
