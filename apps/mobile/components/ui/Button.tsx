import type { TouchableOpacityProps } from "react-native";
import type { VariantProps } from "cva";
import { Children } from "react";
import { ActivityIndicator, Pressable, StyleSheet, View } from "react-native";
import { colors, themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { cn } from "@/lib/utils/classnames";
import { Text } from "./Text";
import { TextStyleContext } from "./text-style-context";
import { buttonVariants } from "./button-variants";

export interface ButtonProps
  extends Omit<TouchableOpacityProps, "disabled">,
    VariantProps<typeof buttonVariants> {
  href?: string;
  isDestructive?: boolean;
}

export const Button = ({
  rounded,
  size,
  loading,
  href,
  className,
  disabled,
  children,
  color = "primary",
  isDestructive,
  fullWidth,
  activeOpacity = 0.7,
  accessibilityState,
  style,
  ...rest
}: ButtonProps) => {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const isDisabled = Boolean(disabled || loading);
  const textColor =
    isDestructive && color === "tertiary"
      ? "danger"
      : color === "primary"
        ? "primaryForeground"
        : color === "invert"
          ? "inverse"
          : "foreground";
  const indicatorColor =
    textColor === "primaryForeground"
      ? colors.primaryForeground
      : textColor === "danger"
        ? resolvedTheme === "dark"
          ? colors.dangerTextDark
          : colors.danger
        : textColor === "inverse"
          ? theme.foregroundInverse
          : theme.foreground;

  return (
    <Pressable
      {...rest}
      accessibilityRole="button"
      accessibilityState={{
        ...accessibilityState,
        disabled: isDisabled,
        busy: Boolean(loading),
      }}
      className={cn(
        buttonVariants({ size, disabled, loading, rounded, color, fullWidth }),
        className,
      )}
      disabled={isDisabled}
      style={({ pressed }) => [style, pressed && { opacity: activeOpacity }]}
    >
      <TextStyleContext.Provider
        value={{ color: textColor, fontWeight: "semibold" }}
      >
        <View style={[styles.content, loading && styles.hiddenContent]}>
          {Children.map(children, (child) =>
            typeof child === "string" || typeof child === "number" ? (
              <Text>{child}</Text>
            ) : (
              child
            ),
          )}
        </View>
      </TextStyleContext.Provider>
      {loading ? (
        <View pointerEvents="none" style={styles.indicator}>
          <ActivityIndicator
            accessible={false}
            size="small"
            color={indicatorColor}
          />
        </View>
      ) : null}
    </Pressable>
  );
};

const styles = StyleSheet.create({
  content: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
  },
  hiddenContent: { opacity: 0 },
  indicator: {
    position: "absolute",
    inset: 0,
    alignItems: "center",
    justifyContent: "center",
  },
});
