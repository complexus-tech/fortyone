import type { GlassIconButtonProps } from "./glass-icon-button.types";
import { View } from "react-native";
import { Button, Host, HStack, Image, RNHostView } from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  contentShape,
  disabled as disabledModifier,
  frame,
  glassEffect,
  shapes,
} from "@expo/ui/swift-ui/modifiers";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export type { GlassIconButtonProps } from "./glass-icon-button.types";

export function GlassIconButton({
  label,
  onPress,
  disabled = false,
  children,
  systemImage,
}: GlassIconButtonProps) {
  const { resolvedTheme } = useTheme();

  return (
    <Host
      matchContents
      colorScheme={resolvedTheme}
      style={{ width: 44, height: 44 }}
    >
      <Button
        onPress={onPress}
        modifiers={[
          accessibilityLabel(label),
          buttonStyle("plain"),
          disabledModifier(disabled),
          glassEffect({
            glass: { variant: "regular", interactive: true },
            shape: "circle",
          }),
        ]}
      >
        <HStack
          modifiers={[
            frame({ width: 44, height: 44 }),
            contentShape(shapes.circle()),
          ]}
        >
          {children ? (
            <RNHostView matchContents>
              <View
                pointerEvents="none"
                style={{
                  width: 44,
                  height: 44,
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                {children}
              </View>
            </RNHostView>
          ) : (
            <Image
              systemName={systemImage}
              size={20}
              color={themeColors[resolvedTheme].foreground}
            />
          )}
        </HStack>
      </Button>
    </Host>
  );
}
