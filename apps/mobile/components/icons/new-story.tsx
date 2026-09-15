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
  const tint = color ?? themeColors[resolvedTheme].icon;

  return (
    <SymbolView
      name="square.and.pencil"
      size={size}
      tintColor={tint}
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
      fallback={
        <Ionicons
          name="create-outline"
          size={size}
          color={tint}
          accessible={false}
        />
      }
    />
  );
}
