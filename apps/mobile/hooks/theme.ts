import { useColorScheme } from "nativewind";
import { useEffect, useState, useCallback } from "react";
import { Appearance } from "react-native";

type Theme = "light" | "dark" | "system";

export function useTheme() {
  const { setColorScheme } = useColorScheme();
  const [theme, setTheme] = useState<Theme>("system"); // 'light' | 'dark' | 'system'
  const [resolvedTheme, setResolvedTheme] = useState(
    Appearance.getColorScheme() || "light"
  );

  useEffect(() => {
    if (theme === "system") {
      // follow system changes
      const listener = Appearance.addChangeListener(({ colorScheme }) => {
        setResolvedTheme(colorScheme || "light");
      });
      setResolvedTheme(Appearance.getColorScheme() || "light");
      setColorScheme("system");
      return () => listener.remove();
    } else {
      // fixed theme
      setColorScheme(theme);
      setResolvedTheme(theme);
    }
  }, [theme, setColorScheme]);

  const changeTheme = useCallback((newTheme: Theme) => {
    setTheme(newTheme);
  }, []);

  return {
    theme, // current user preference
    resolvedTheme, // actual active theme
    setTheme: changeTheme,
  };
}
