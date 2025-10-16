import { useColorScheme } from "nativewind";
import { useState, useEffect, useCallback } from "react";
import { MMKV } from "react-native-mmkv";

type Theme = "light" | "dark" | "system";

const storage = new MMKV();
const THEME_STORAGE_KEY = "app-theme";

export function useTheme() {
  const { colorScheme, setColorScheme } = useColorScheme();
  const [theme, setTheme] = useState<Theme>("system");

  // Load theme from storage on mount
  useEffect(() => {
    const loadTheme = () => {
      const storedTheme = storage.getString(THEME_STORAGE_KEY);
      if (storedTheme && ["light", "dark", "system"].includes(storedTheme)) {
        const savedTheme = storedTheme as Theme;
        setTheme(savedTheme);
        setColorScheme(savedTheme);
      }
    };

    loadTheme();
  }, [setColorScheme]);

  // Update resolved theme when system theme changes
  useEffect(() => {
    if (theme === "system") {
      setColorScheme("system");
    }
  }, [theme, setColorScheme]);

  const changeTheme = useCallback(
    (newTheme: Theme) => {
      setTheme(newTheme);
      setColorScheme(newTheme);
      storage.set(THEME_STORAGE_KEY, newTheme);
    },
    [setColorScheme]
  );

  return {
    theme,
    resolvedTheme: colorScheme,
    setTheme: changeTheme,
  };
}
