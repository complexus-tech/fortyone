import type { Config } from "tailwindcss";

// each package will define it's own content
const config: Omit<Config, "content"> = {
  theme: {
    colors: {
      transparent: "transparent",
      current: "currentColor",
      primary: "#22c55e",
      // primary: "#EA6060",
      secondary: "#002F61",
      black: "#0c0c0c",
      white: "#ffffff",
      success: "#22c55e",
      warning: "#eab308",
      danger: "#f43f5e",
      info: "#06b6d4",
      gray: {
        DEFAULT: "#44403c",
        50: "#f5f5f4",
        100: "#e7e5e4",
        200: "#d6d3d1",
        250: "#4b5563",
        300: "#a3a3a3",
      },
      dark: {
        DEFAULT: "#131313",
        300: "#181818",
        200: "#1E1E1E",
        100: "#262626",
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
