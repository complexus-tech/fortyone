import type { Config } from "tailwindcss";
import sharedConfig from "tailwind-config/tailwind.config";

const config: Pick<
  Config,
  "content" | "presets" | "darkMode" | "plugins" | "theme"
> = {
  darkMode: "class",
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/ui/src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    colors: {
      transparent: "transparent",
      current: "currentColor",
      primary: "#7FBD42",
      secondary: "#F15C27",
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
  },
  presets: [sharedConfig],
  plugins: [require("@tailwindcss/typography")],
};

export default config;
