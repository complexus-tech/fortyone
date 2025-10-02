import React from "react";
import { ScrollView } from "react-native";
import { useRouter } from "expo-router";
import { SafeContainer, Section } from "@/components/ui";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { Team } from "./components/team";

export const Home = () => {
  const router = useRouter();

  const handleTeamPress = (teamId: string) => {
    router.push(`/team/${teamId}/stories`);
  };

  return (
    <SafeContainer>
      <Header />
      <ScrollView className="flex-1">
        <Overview />
        <Section title="Your Teams">
          <Team
            name="Engineering"
            color="#FF9500"
            onPress={() => handleTeamPress("1")}
          />
          <Team
            name="Complexus"
            color="#34C759"
            onPress={() => handleTeamPress("2")}
          />
        </Section>
      </ScrollView>
    </SafeContainer>
  );
};
