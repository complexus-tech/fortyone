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

        <TextInput
          placeholder="Search"
          className="border rounded-xl px-4 h-14 mx-4 mb-4"
        />
        <TextInput
          placeholder="Search"
          className="border rounded-xl px-4 h-14 mx-4"
        />
        <Button className="mt-6 mx-4">Save changes</Button>

        <ProfileForm />
      </ScrollView>
    </SafeContainer>
  );
};
