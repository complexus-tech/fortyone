import React from "react";
import { Linking, Pressable } from "react-native";
import { Link } from "@/types/link";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { toast } from "sonner-native";

type CardProps = {
  link: Link;
};

export const Card = ({ link }: CardProps) => {
  const { resolvedTheme } = useTheme();
  const handlePress = async () => {
    try {
      const canOpen = await Linking.canOpenURL(link.url);
      if (!canOpen) {
        toast.error("Could not open this link", {
          description: "No app is available to open this link.",
        });
        return;
      }
      await Linking.openURL(link.url);
    } catch (error) {
      toast.error("Could not open this link", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
  };

  return (
    <Pressable
      accessibilityRole="link"
      accessibilityLabel={link.title || link.url}
      className="active:bg-gray-50 dark:active:bg-dark py-4"
      onPress={handlePress}
    >
      <Row asContainer align="center" className="flex-1 gap-1.5">
        <SymbolView
          name="globe"
          size={20}
          weight="semibold"
          tintColor={themeColors[resolvedTheme].textMuted}
        />
        <Text numberOfLines={1} className="flex-1">
          {link.title || link.url}
        </Text>
      </Row>
    </Pressable>
  );
};
