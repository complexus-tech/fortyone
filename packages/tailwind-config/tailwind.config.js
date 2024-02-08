/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/**/*.{js,ts,jsx,tsx,mdx}',
    '../../packages/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    fontFamily: {
      body: 'Figtree, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Roboto, Arial, sans-serif',
      heading:
        'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Roboto, Arial, sans-serif',
    },
    colors: {
      white: '#ffffff',
      transparent: 'transparent',
      current: 'currentColor',
      primary: '#EA6060',
      secondary: '#002F61',
      black: '#1D1D1F',
      gray: '#A1A1A6',
      white: '#ffffff',
      success: '#22c55e',
      warning: '#eab308',
      danger: '#f43f5e',
      info: '#06b6d4',
      gray: {
        DEFAULT: '#A1A1A6',
        50: '#f3f4f6',
        100: '#e2e8f0',
        200: '#cbd5e1',
        250: '#4b5563',
        300: '#27272a',
      },
      dark: {
        DEFAULT: '#0f172a',
        50: '#475569',
        100: '#334155',
        200: '#1e293b',
        300: '#020617',
      },
    },
    extend: {
      theme: {
        extend: {
          screens: {
            '3xl': '1900px',
          },
        },
      },
    },
  },
};
