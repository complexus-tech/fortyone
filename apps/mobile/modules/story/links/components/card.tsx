import React from "react";
import { Linking, Pressable } from "react-native";
import { Link } from "@/types/link";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";

type CardProps = {
  link: Link;
};

export const Card = ({ link }: CardProps) => {
  const { colorScheme } = useColorScheme();
  const handlePress = async () => {
    const canOpen = await Linking.canOpenURL(link.url);
    if (canOpen) {
      await Linking.openURL(link.url);
    }
  };

  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-300 border-t border-gray-50 dark:border-dark"
      onPress={handlePress}
    >
      <Row align="center" justify="between" className="p-4" gap={3}>
        <Row align="center" gap={2} className="flex-1">
          <SymbolView
            name="globe"
            size={20}
            weight="semibold"
            tintColor={
              colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
            }
          />
          <Row align="center" className="flex-1" gap={2}>
            <Text numberOfLines={1} className="flex-1">
              {link.title || link.url}
            </Text>
            <SymbolView
              name="arrow.up.right"
              size={12}
              weight="semibold"
              tintColor={
                colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
              }
            />
          </Row>
        </Row>
      </Row>
    </Pressable>
  );
};
