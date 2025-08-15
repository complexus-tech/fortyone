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
      gray: {
        DEFAULT: "#6B665C",
        50: "#F8F6F2",
        100: "#ECE9E4",
        200: "#DAD6D0",
        250: "#5E5A52",
        300: "#A19B94",
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
