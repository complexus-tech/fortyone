import React from "react";
import { SafeContainer } from "@/components/ui";
import { ScrollView } from "react-native";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import { Form } from "./components/form";
import { Header } from "./components/header";

export const Settings = () => {
  const { resolvedTheme } = useTheme();
  return (
    <SafeContainer isFull>
      <ScrollView
        style={{
          paddingTop: 44,
          paddingHorizontal: 16,
          flex: 1,
          backgroundColor:
            resolvedTheme === "light" ? colors.white : colors.black,
        }}
        contentContainerStyle={{ paddingBottom: 120 }}
      >
        <Header />
        <Form />
      </ScrollView>
    </SafeContainer>
  );
};
