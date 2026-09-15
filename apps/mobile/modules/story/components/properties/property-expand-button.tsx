import { Pressable, StyleSheet, View } from "react-native";
import Svg, { Path } from "react-native-svg";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

type PropertyExpandButtonProps = {
  expanded: boolean;
  onPress: () => void;
};

export function PropertyExpandButton({
  expanded,
  onPress,
}: PropertyExpandButtonProps) {
  const { resolvedTheme } = useTheme();

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={
        expanded ? "Show fewer properties" : "Show all properties"
      }
      accessibilityState={{ expanded }}
      onPress={onPress}
      style={({ pressed }) => [styles.button, pressed && styles.pressed]}
    >
      <View pointerEvents="none" style={styles.icon}>
        <Svg
          width={18}
          height={18}
          viewBox="0 0 18 18"
          fill="none"
          accessible={false}
          accessibilityElementsHidden
          importantForAccessibility="no"
        >
          <Path
            d={expanded ? "M3 9h12" : "M3 9h12M9 3v12"}
            stroke={themeColors[resolvedTheme].textMuted}
            strokeWidth={1.8}
            strokeLinecap="round"
          />
        </Svg>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    width: 44,
    height: 44,
    flexShrink: 0,
    alignItems: "flex-start",
    justifyContent: "center",
  },
  icon: {
    width: 28,
    height: 28,
    alignItems: "center",
    justifyContent: "center",
  },
  pressed: { opacity: 0.6 },
});
