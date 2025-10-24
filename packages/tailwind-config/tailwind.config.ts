import type { Config } from "tailwindcss";

// each package will define it's own content
const config: Omit<Config, "content"> = {
  theme: {
    fontFamily: {
      satoshi: ["var(--font-satoshi)"],
      inter: ["var(--font-inter)"],
      sans: ["var(--font-sans)"],
    },
    colors: {
      transparent: "transparent",
      current: "currentColor",
      primary: "#f43f5e",
      secondary: "#002F61",
      black: "#08090a",
      white: "#ffffff",
      success: "#22c55e",
      warning: "#eab308",
      danger: "#f43f5e",
      info: "#06b6d4",
      sidebar: "#FCF5E5",
      light: "#fffff0",

      // --color-gray: #5f5f5f;
      // --color-gray-50: #f7f7f7;
      // --color-gray-100: #e8e8e8;
      // --color-gray-200: #d1d1d1;
      // --color-gray-300: #9e9e9e;
      // --color-gray-400: #7d7d7d;
      gray: {
        DEFAULT: "#5f5f5f",
        50: "#f7f7f7",
        100: "#e8e8e8",
        200: "#d1d1d1",
        300: "#9e9e9e",
        400: "#7d7d7d",
      },
      dark: {
        DEFAULT: "#0c0d0e",
        300: "#101214",
        200: "#16181a",
        100: "#1d1f22",
        50: "#25282b",
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
