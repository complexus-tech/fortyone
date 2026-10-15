import type { Settings } from "./types";
export function applyTheme(theme: Settings["theme"]) {
  document.documentElement.dataset.theme = theme;
}
