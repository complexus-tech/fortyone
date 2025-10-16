import React from "react";
import { SafeContainer } from "@/components/ui";
import { ScrollView } from "react-native";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";
import { ProfileForm } from "./components/profile-form";
import { Header } from "./components/header";

export const Account = () => {
  const { colorScheme } = useColorScheme();
  return (
    <SafeContainer isFull>
      <ScrollView
        style={{
          paddingTop: 44,
          flex: 1,
          backgroundColor:
            colorScheme === "light" ? colors.white : colors.dark[200],
        }}
        contentContainerStyle={{ paddingBottom: 120 }}
      >
        <Header />
        <ProfileForm />
      </ScrollView>
    </SafeContainer>
  );
};
