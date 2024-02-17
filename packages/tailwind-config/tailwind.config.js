/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    fontFamily: {
      body: "Figtree, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Roboto, Arial, sans-serif",
      heading:
        "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Roboto, Arial, sans-serif",
    },
    colors: {
      white: "#ffffff",
      transparent: "transparent",
      current: "currentColor",
      primary: "#EA6060",
      secondary: "#002F61",
      black: "#1D1D1F",
      gray: "#A1A1A6",
      white: "#ffffff",
      success: "#22c55e",
      warning: "#eab308",
      danger: "#f43f5e",
      info: "#06b6d4",
      gray: {
        DEFAULT: "#A1A1A6",
        50: "#f3f4f6",
        100: "#e5e7eb",
        200: "#d1d5db",
        250: "#9ca3af",
        300: "#27272a",
      },
      dark: {
        DEFAULT: "#000000",
        300: "#101010",
        200: "#1c1c1c",
        100: "#2b2b2b",
        50: "#444444",
      },
    },
    extend: {
      theme: {
        extend: {
          screens: {
            "3xl": "1900px",
          },
        },
      },
    },
  },
};
