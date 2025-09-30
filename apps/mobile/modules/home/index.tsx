import React from "react";
import { ScrollView } from "react-native";
import { SafeContainer, Section } from "@/components/ui";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { Team } from "./components/team";

export const Home = () => {
  const stats = {
    closed: 12,
    overdue: 3,
    inProgress: 8,
    created: 24,
    assigned: 15,
  };

  return (
    <SafeContainer>
      <Header />
      <ScrollView className="flex-1">
        <Overview stats={stats} />
        <Section title="Your Teams">
          <Team name="Engineering" color="#FF9500" />
          <Team name="Complexus" color="#34C759" />
        </Section>
      </ScrollView>
    </SafeContainer>
  );
};
