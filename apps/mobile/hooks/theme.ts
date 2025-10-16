import { useColorScheme } from "nativewind";
import { useState, useCallback } from "react";

type Theme = "light" | "dark" | "system";

export function useTheme() {
  const { colorScheme, setColorScheme } = useColorScheme();
  const [theme, setTheme] = useState<Theme>("system");

  const changeTheme = useCallback(
    (newTheme: Theme) => {
      setTheme(newTheme);
      setColorScheme(newTheme);
    },
    [setColorScheme]
  );

  return {
    theme,
    resolvedTheme: colorScheme,
    setTheme: changeTheme,
  };
}
