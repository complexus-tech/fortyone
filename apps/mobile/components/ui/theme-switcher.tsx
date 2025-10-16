import React from "react";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
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
    <HStack spacing={8} onPress={onPress}>
      <Image
        systemName={theme.icon}
        color={resolvedTheme === "light" ? "black" : "white"}
        size={20}
        modifiers={[frame({ width: 30, height: 30 })]}
      />
      <VStack alignment="leading">
        <Text lineLimit={1} size={15} weight="medium">
          {theme.label}
        </Text>
      </VStack>
      <Spacer />
      {isActive && (
        <Image
          systemName="checkmark.circle.fill"
          color={resolvedTheme === "light" ? "black" : "white"}
          size={18}
        />
      )}
    </HStack>
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
    <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
      <HStack>
        <Text
          weight="medium"
          color={
            resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
          size={14}
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
