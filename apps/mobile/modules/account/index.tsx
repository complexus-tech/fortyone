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
          placeholder="Enter name"
          className="border rounded-[0.7rem] border-gray-100 bg-gray-50 px-4 h-14 mx-4 mb-4 dark:border-dark-50 dark:bg-dark-300 dark:text-white"
        />
        <TextInput
          placeholder="Enter username"
          className="border rounded-[0.7rem] border-gray-100 bg-gray-50 px-4 h-14 mx-4 dark:border-dark-50 dark:bg-dark-300 dark:text-white"
        />
        <Button className="mt-5 mx-4">Save changes</Button>

        {/* <ProfileForm /> */}
      </ScrollView>
    </SafeContainer>
  );
};
