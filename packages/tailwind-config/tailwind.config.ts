import type { Config } from "tailwindcss";

// each package will define it's own content
const config: Omit<Config, "content"> = {
  theme: {
    fontFamily: {
      satoshi: ["var(--font-satoshi)"],
      inter: ["var(--font-inter)"],
    },
    colors: {
      transparent: "transparent",
      current: "currentColor",
      primary: "#EA6060",
      secondary: "#002F61",
      black: "#010101",
      white: "#ffffff",
      success: "#22c55e",
      warning: "#eab308",
      danger: "#f43f5e",
      info: "#06b6d4",
      sidebar: "#FCF5E5",
      light: "#fffff0",
      // gray: {
      //   DEFAULT: "#44403c",
      //   50: "#f5f5f4",
      //   100: "#e7e5e4",
      //   200: "#d6d3d1",
      //   250: "#4b5563",
      //   300: "#a3a3a3",
      // },
      gray: {
        DEFAULT: "#6B665C", // A softer, warm gray
        50: "#F8F6F2", // Very light warm gray, almost off-white
        100: "#ECE9E4", // Light warm gray
        200: "#DAD6D0", // Medium light warm gray
        250: "#5E5A52", // A slightly darker warm gray
        300: "#A19B94",
      },
      dark: {
        DEFAULT: "#0b0b0b",
        300: "#161616",
        200: "#1b1b1b",
        100: "#232323",
        50: "#2F2F2F",
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
