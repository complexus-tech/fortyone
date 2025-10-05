import React from "react";
import { View } from "react-native";
import { Title } from "./components/title";
import { Properties } from "./components/properties";

export const Story = () => {
  return (
    <View className="flex-1">
      <Title />
      <Properties />
    </View>
  );
};
