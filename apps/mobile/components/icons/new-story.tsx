import { Ionicons } from "@expo/vector-icons";
import { SymbolView } from "expo-symbols";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

export function NewStoryIcon({
  size = 20,
  color,
}: {
  size?: number;
  color?: string;
}) {
  const { resolvedTheme } = useTheme();
  const tint = color ?? themeColors[resolvedTheme].foreground;
  return (
    <SymbolView
      name="plus"
      weight="semibold"
      size={size}
      tintColor={tint}
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
      fallback={
        <Ionicons name="add" size={size} color={tint} accessible={false} />
      }
    />
  );
}
