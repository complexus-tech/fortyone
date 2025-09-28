import React from "react";
import { View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Host, Button } from "@expo/ui/swift-ui";
import { Text } from "@/components/ui";

interface HeaderProps {
  title: string;
  onSettingsPress?: () => void;
  showSettingsIcon?: boolean;
}

export const Header = ({
  title,
  onSettingsPress,
  showSettingsIcon = true,
}: HeaderProps) => {
  const insets = useSafeAreaInsets();

  return (
    <View className="bg-white mb-4" style={{ paddingTop: insets.top }}>
      <View className="flex-row justify-between items-start py-2">
        <View className="flex-row items-center">
          <Text fontSize="2xl" fontWeight="semibold" color="black">
            {title}
          </Text>
        </View>
        <Host matchContents>
          <Button
            onPress={() => {
              alert("New Story");
            }}
            systemImage="plus"
            variant="glass"
          ></Button>
        </Host>
      </View>
    </View>
  );
};
