/**
 * Matches the Tailwind configuration
 */

// --color-dark: #0a0a0a;
// --color-dark-50: #3f3f3f;
// --color-dark-100: #262626;
// --color-dark-200: #171717;
// --color-dark-300: #0f0f0f;
export const colors = {
  primary: "#f43f5e",
  secondary: "#002f61",
  black: "#000000",
  white: "#ffffff",
  transparent: "transparent",
  current: "currentColor",
  success: "#22c55e",
  warning: "#eab308",
  danger: "#f43f5e",
  info: "#06b6d4",
  sidebar: "#FCF5E5",
  light: "#fffff0",
  gray: {
    DEFAULT: "#5f5f5f",
    50: "#f7f7f7",
    100: "#e8e8e8",
    200: "#d1d1d1",
    300: "#9e9e9e",
    400: "#7d7d7d",
  },
  dark: {
    DEFAULT: "#0a0a0a",
    50: "#3f3f3f",
    100: "#262626",
    200: "#171717",
    300: "#0f0f0f",
  },
} as const;
