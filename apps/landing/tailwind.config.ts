import type { Config } from "tailwindcss";
import sharedConfig from "tailwind-config/tailwind.config";

const config: Pick<Config, "content" | "presets" | "darkMode" | "plugins"> = {
  darkMode: "media",
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/ui/src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  presets: [sharedConfig],
};

export default config;
