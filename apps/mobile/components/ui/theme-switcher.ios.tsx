import React from "react";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { themeColors } from "@/constants/colors";
import { Button, HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import {
  frame,
  font,
  foregroundStyle,
  lineLimit,
  buttonStyle,
  accessibilityLabel,
} from "@expo/ui/swift-ui/modifiers";
import { SFSymbol } from "expo-symbols";
import { useTheme } from "@/hooks";

const ThemeItem = ({
  isActive,
  theme,
  onPress,
}: {
  isActive: boolean;
  theme: {
    label: string;
    value: string;
    icon: SFSymbol;
  };
  onPress: () => void;
}) => {
  const { resolvedTheme } = useTheme();
  return (
    <Button
      onPress={onPress}
      modifiers={[
        buttonStyle("plain"),
        accessibilityLabel(`${theme.label}${isActive ? ", selected" : ""}`),
      ]}
    >
      <HStack spacing={8}>
        <Image
          systemName={theme.icon}
          color={themeColors[resolvedTheme].foreground}
          size={18}
          modifiers={[frame({ width: 28, height: 28 })]}
        />
        <VStack alignment="leading">
          <Text modifiers={[lineLimit(1), font({ size: 15 })]}>
            {theme.label}
          </Text>
        </VStack>
        <Spacer />
        {isActive && (
          <Image
            systemName="checkmark.circle.fill"
            color={themeColors[resolvedTheme].foreground}
            size={18}
          />
        )}
      </HStack>
    </Button>
  );
};

export const ThemeSwitcher = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  const { theme: currentTheme, resolvedTheme, setTheme } = useTheme();

  const themes: {
    label: string;
    value: "light" | "dark" | "system";
    icon: SFSymbol;
  }[] = [
    {
      label: "Light",
      value: "light",
      icon: "sun.max.fill",
    },
    {
      label: "Dark",
      value: "dark",
      icon: "moon.fill",
    },
    {
      label: "Automatic",
      value: "system",
      icon: "platter.filled.top.iphone",
    },
  ];

  return (
    <BottomSheetModal
      nativeContent
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
    >
      <HStack>
        <Text
          modifiers={[
            font({ size: 14, weight: "medium" }),
            foregroundStyle(themeColors[resolvedTheme].textMuted),
          ]}
        >
          Appearance
        </Text>
      </HStack>
      {themes.map((theme) => (
        <ThemeItem
          key={theme.value}
          isActive={theme.value === currentTheme}
          theme={theme}
          onPress={() => {
            setTheme(theme.value);
            setIsOpened(false);
          }}
        />
      ))}
    </BottomSheetModal>
  );
};
