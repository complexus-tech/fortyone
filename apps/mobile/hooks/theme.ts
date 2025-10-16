import { useColorScheme } from "nativewind";
import { useState } from "react";

type Theme = "light" | "dark" | "system";

export function useTheme() {
  const { colorScheme, setColorScheme } = useColorScheme();
  const [theme, setTheme] = useState<Theme>("system");

  const changeTheme = (newTheme: Theme) => {
    setTheme(newTheme);
    setColorScheme(newTheme);
  };

  return {
    theme,
    resolvedTheme: colorScheme,
    setTheme: changeTheme,
  };
}
