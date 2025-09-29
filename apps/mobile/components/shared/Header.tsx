import React from "react";
import { Pressable, View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { SymbolView } from "expo-symbols";
import { Avatar, Text } from "@/components/ui";
import { useRouter } from "expo-router";

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
  const router = useRouter();

  return (
    <View
      className="mb-5 flex-row justify-between items-center"
      style={{ paddingTop: insets.top }}
    >
      <View className="flex-row items-center py-2 gap-2">
        <Avatar
          name="John Doe"
          size="md"
          src="https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s576-c-no"
        />
        <Text fontSize="2xl" fontWeight="semibold" numberOfLines={1}>
          {title}
        </Text>
      </View>
      {showSettingsIcon && (
        <Pressable
          onPress={() => {
            router.push("/settings");
          }}
        >
          <SymbolView name="gear" size={28} tintColor="#444" />
        </Pressable>
      )}
    </View>
  );
};
