import React from "react";
import { Linking, Pressable } from "react-native";
import { Link } from "@/types/link";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";

type CardProps = {
  link: Link;
};

export const Card = ({ link }: CardProps) => {
  const { resolvedTheme } = useTheme();
  const handlePress = async () => {
    const canOpen = await Linking.canOpenURL(link.url);
    if (canOpen) {
      await Linking.openURL(link.url);
    }
  };

  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark py-4"
      onPress={handlePress}
    >
      <Row asContainer align="center" className="flex-1 gap-1.5">
        <SymbolView
          name="globe"
          size={20}
          weight="semibold"
          tintColor={
            resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
        <Text numberOfLines={1} className="flex-1">
          {link.title || link.url}
        </Text>
      </Row>
    </Pressable>
  );
};
