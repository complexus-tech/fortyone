import { useState } from "react";

type Theme = "light" | "dark";

const getTheme = (): Theme => {
  if (typeof window === "undefined") return "light";

  if (
    window.localStorage.theme === "dark" ||
    (!("theme" in window.localStorage) &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
  ) {
    return "dark";
  }

  return "light";
};

const applyTheme = (theme: Theme) => {
  if (typeof document === "undefined") return;

  if (theme === "dark") {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
};

export const useTheme = () => {
  const [theme, setThemeState] = useState<Theme>(() => {
    const initialTheme = getTheme();
    applyTheme(initialTheme);
    return initialTheme;
  });

  const toggleTheme = () => {
    const nextTheme = theme === "light" ? "dark" : "light";
    setThemeState(nextTheme);
    window.localStorage.setItem("theme", nextTheme);
    applyTheme(nextTheme);
  };

  return { theme, toggleTheme };
};
