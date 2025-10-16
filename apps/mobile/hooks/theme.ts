import { useColorScheme } from "nativewind";
import { useEffect, useCallback } from "react";
import { useMMKVString } from "react-native-mmkv";

type Theme = "light" | "dark" | "system";

const THEME_STORAGE_KEY = "app-theme";

export function useTheme() {
  const { colorScheme, setColorScheme } = useColorScheme();
  const [theme, setTheme] = useMMKVString(THEME_STORAGE_KEY);

  useEffect(() => {
    if (theme === "system") {
      setColorScheme("system");
    }
  }, [theme, setColorScheme]);

  const changeTheme = useCallback(
    (newTheme: Theme) => {
      setTheme(newTheme);
      setColorScheme(newTheme);
    },
    [setColorScheme, setTheme]
  );

  return {
    theme: (theme || "system") as Theme,
    resolvedTheme: colorScheme,
    setTheme: changeTheme,
  };
}
