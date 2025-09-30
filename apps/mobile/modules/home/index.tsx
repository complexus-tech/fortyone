import React from "react";
import { ScrollView } from "react-native";
import { Header } from "./components/header";
import { SafeContainer, Section } from "@/components/ui";
import { Team } from "./components/team";

export const Home = () => {
  return (
    <SafeContainer>
      <Header />
      <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
        <Section title="Your Teams">
          <Team name="Engineering" color="#FF9500" />
          <Team name="Complexus" color="#34C759" />
        </Section>
      </ScrollView>
    </SafeContainer>
  );
};
