import React from "react";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { colors } from "@/constants";
import { Pressable, View, Text as RNText } from "react-native";
import { Ionicons } from "@expo/vector-icons";
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
    icon: string;
  };
  onPress: () => void;
}) => {
  const { resolvedTheme } = useTheme();
  return (
    <Pressable
      onPress={onPress}
      style={{
        flexDirection: "row",
        alignItems: "center",
        paddingVertical: 12,
        paddingHorizontal: 16,
        gap: 8,
      }}
    >
      <View
        style={{
          width: 30,
          height: 30,
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Ionicons
          name={theme.icon}
          size={20}
          color={resolvedTheme === "light" ? "black" : "white"}
        />
      </View>
      <View style={{ flex: 1 }}>
        <RNText
          style={{
            fontSize: 16,
            color: resolvedTheme === "light" ? "black" : "white",
          }}
        >
          {theme.label}
        </RNText>
      </View>
      {isActive && (
        <Ionicons
          name="checkmark-circle"
          size={18}
          color={resolvedTheme === "light" ? "black" : "white"}
        />
      )}
    </Pressable>
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
    icon: string;
  }[] = [
    {
      label: "Light",
      value: "light",
      icon: "sunny",
    },
    {
      label: "Dark",
      value: "dark",
      icon: "moon",
    },
    {
      label: "Automatic",
      value: "system",
      icon: "phone-portrait",
    },
  ];

  return (
    <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
      <View style={{ paddingHorizontal: 16, paddingBottom: 16 }}>
        <RNText
          style={{
            fontWeight: "500",
            color:
              resolvedTheme === "light"
                ? colors.gray.DEFAULT
                : colors.gray[300],
            fontSize: 14,
            marginBottom: 8,
          }}
        >
          Appearance
        </RNText>
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
      </View>
    </BottomSheetModal>
  );
};
