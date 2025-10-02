import React from "react";
import { ScrollView } from "react-native";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { Teams } from "./components/teams";

export const Home = () => {
  return (
    <SafeContainer isFull>
      <Header />
      <ScrollView className="flex-1">
        <Overview />
        <Teams />
      </ScrollView>
    </SafeContainer>
  );
};
