import { useColorScheme } from "nativewind";
import { useEffect, useState, useCallback } from "react";

type Theme = "light" | "dark" | "system";

export function useTheme() {
  const { colorScheme, setColorScheme } = useColorScheme();
  const [theme, setTheme] = useState<Theme>("system");
  const [resolvedTheme, setResolvedTheme] = useState<"light" | "dark">(
    colorScheme || "light"
  );

  useEffect(() => {
    if (theme === "system") {
      setResolvedTheme(colorScheme);
      setColorScheme("system");
    } else {
      setColorScheme(theme);
      setResolvedTheme(theme);
    }
  }, [theme, colorScheme, setColorScheme]);

  const changeTheme = useCallback((newTheme: Theme) => {
    setTheme(newTheme);
  }, []);

  return {
    theme,
    resolvedTheme,
    setTheme: changeTheme,
  };
}
