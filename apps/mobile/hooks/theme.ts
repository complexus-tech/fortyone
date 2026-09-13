import { Appearance, useColorScheme } from "react-native";
import { useEffect, useCallback } from "react";
import { useMMKVString } from "react-native-mmkv";

type Theme = "light" | "dark" | "system";

const THEME_STORAGE_KEY = "app-theme";

export function useTheme() {
  const colorScheme = useColorScheme();
  const [theme, setTheme] = useMMKVString(THEME_STORAGE_KEY);

  useEffect(() => {
    Appearance.setColorScheme(
      theme === "light" || theme === "dark" ? theme : "unspecified",
    );
  }, [theme]);

  const changeTheme = useCallback(
    (newTheme: Theme) => {
      setTheme(newTheme);
      Appearance.setColorScheme(
        newTheme === "system" ? "unspecified" : newTheme,
      );
    },
    [setTheme],
  );

  return {
    theme: theme === "light" || theme === "dark" ? theme : "system",
    resolvedTheme:
      colorScheme === "dark" ? ("dark" as const) : ("light" as const),
    setTheme: changeTheme,
  };
}
