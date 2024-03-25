import type { Config } from "tailwindcss";

// each package will define it's own content
const config: Omit<Config, "content"> = {
  theme: {
    colors: {
      transparent: "transparent",
      current: "currentColor",
      primary: "#EA6060",
      secondary: "#002F61",
      black: "#1D1D1F",
      white: "#ffffff",
      success: "#22c55e",
      warning: "#eab308",
      danger: "#f43f5e",
      info: "#06b6d4",
      gray: {
        DEFAULT: "#A1A1A6",
        50: "#f3f4f6",
        100: "#e2e8f0",
        200: "#cbd5e1",
        250: "#4b5563",
        300: "#27272a",
      },
      dark: {
        DEFAULT: "#181818",
        50: "#383838",
        100: "#2b2b2b",
        200: "#282828",
        300: "#1f1f1f",
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

export default config;
