import React from "react";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";
import { SFSymbol } from "expo-symbols";

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
  const { colorScheme } = useColorScheme();

  return (
    <HStack spacing={8} onPress={onPress}>
      <Image
        systemName={theme.icon}
        color={colorScheme === "light" ? "black" : "white"}
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
          color={colorScheme === "light" ? "black" : "white"}
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
  const { colorScheme, setColorScheme } = useColorScheme();

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
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
          size={14}
        >
          Appearance
        </Text>
      </HStack>
      {themes.map((theme) => (
        <ThemeItem
          key={theme.value}
          isActive={theme.value === colorScheme}
          theme={theme}
          onPress={() => setColorScheme(theme.value)}
        />
      ))}
    </BottomSheetModal>
  );
};
