import type { Config } from "tailwindcss";
import sharedConfig from "tailwind-config/tailwind.config";

const config: Pick<
  Config,
  "content" | "presets" | "darkMode" | "plugins" | "theme"
> = {
  darkMode: "class",
  theme: {
    fontFamily: {
      heading: ["var(--font-heading)"],
      body: [
        "-apple-system",
        "BlinkMacSystemFont",
        "var(--font-body)",
        "Segoe UI",
        "sans-serif",
      ],
    },
  },
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/ui/src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  presets: [sharedConfig],
  plugins: [require("@tailwindcss/typography")],
};

export default config;
