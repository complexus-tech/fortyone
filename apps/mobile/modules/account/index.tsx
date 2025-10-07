import React from "react";
import { Avatar, Col, Row, SafeContainer, Text, Button } from "@/components/ui";
import { ScrollView, TextInput } from "react-native";
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
          paddingTop: 36,
          flex: 1,
          backgroundColor:
            colorScheme === "light" ? colors.white : colors.dark.DEFAULT,
        }}
        contentContainerStyle={{ paddingBottom: 120 }}
      >
        <Header />
        <ProfileForm />
      </ScrollView>
    </SafeContainer>
  );
};
