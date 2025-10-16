import { useColorScheme } from "nativewind";
import { useState, useEffect, useCallback } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";

type Theme = "light" | "dark" | "system";

const THEME_STORAGE_KEY = "app-theme";

export function useTheme() {
  const { colorScheme, setColorScheme } = useColorScheme();
  const [theme, setTheme] = useState<Theme>("system");

  // Load theme from storage on mount
  useEffect(() => {
    const loadTheme = async () => {
      const storedTheme = await AsyncStorage.getItem(THEME_STORAGE_KEY);
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
    async (newTheme: Theme) => {
      setTheme(newTheme);
      setColorScheme(newTheme);
      try {
        await AsyncStorage.setItem(THEME_STORAGE_KEY, newTheme);
      } catch (error) {
        console.error("Failed to save theme:", error);
      }
    },
    [setColorScheme]
  );

  return {
    theme,
    resolvedTheme: colorScheme,
    setTheme: changeTheme,
  };
}
