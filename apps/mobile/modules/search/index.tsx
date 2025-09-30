import React from "react";
import { View } from "react-native";
import { SafeContainer, Text } from "@/components/ui";
import { Header } from "./components/header";

export const Search = () => {
  return (
    <SafeContainer>
      <Header />
      <View className="flex-1 justify-center items-center">
        <Text fontSize="lg" color="muted">
          Search functionality coming soon
        </Text>
      </View>
    </SafeContainer>
  );
};
