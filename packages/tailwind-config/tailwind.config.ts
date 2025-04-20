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
      primary: "#EA6060",
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
        // DEFAULT: "#09090a",
        // 300: "#0d0e10",
        // 200: "#131517",
        // 100: "#191b1e",
        // 50: "#21232a",
        // DEFAULT: "#050506",
        // 300: "#08090b",
        // 200: "#0e0f12",
        // 100: "#141619",
        // 50: "#1b1d23",
        // DEFAULT: "#050507",
        // 300: "#09090c",
        // 200: "#0d0d12",
        // 100: "#12131a",
        // 50: "#191a24",
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
