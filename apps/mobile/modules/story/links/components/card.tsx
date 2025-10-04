import React from "react";
import { Linking, Pressable } from "react-native";
import { Link } from "@/types/link";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

type CardProps = {
  link: Link;
};

export const Card = ({ link }: CardProps) => {
  const handlePress = async () => {
    const canOpen = await Linking.canOpenURL(link.url);
    if (canOpen) {
      await Linking.openURL(link.url);
    }
  };

  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-300"
      onPress={handlePress}
    >
      <Row
        align="center"
        justify="between"
        className="px-4 py-3 border-t-[0.5px] border-gray-100 dark:border-dark-100"
        gap={3}
      >
        <Row align="center" gap={3} className="flex-1">
          <SymbolView name="link" size={20} tintColor={colors.primary} />
          <Row className="flex-1" gap={2}>
            <Text numberOfLines={1} className="flex-1">
              {link.title || link.url}
            </Text>
            <SymbolView
              name="arrow.up.right"
              size={16}
              tintColor={colors.gray.DEFAULT}
            />
          </Row>
        </Row>
      </Row>
    </Pressable>
  );
};
