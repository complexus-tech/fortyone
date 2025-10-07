import React from "react";
import { SafeContainer, Text } from "@/components/ui";
import { ScrollView } from "react-native";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";

export const Account = () => {
  const { colorScheme } = useColorScheme();
  return (
    <SafeContainer>
      <ScrollView
        style={{
          paddingTop: 36,
          flex: 1,
          backgroundColor:
            colorScheme === "light" ? colors.white : colors.dark[300],
        }}
      >
        <Text>Account</Text>
      </ScrollView>
    </SafeContainer>
  );
};
