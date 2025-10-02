import React from "react";
import { ScrollView, ActivityIndicator } from "react-native";
import { useRouter } from "expo-router";
import { SafeContainer, Section, Text } from "@/components/ui";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { Team } from "./components/team";
import { useOverviewStats } from "./hooks/use-overview-stats";

export const Home = () => {
  const router = useRouter();
  const { data: stats, isLoading: statsLoading } = useOverviewStats();

  const handleTeamPress = (teamId: string) => {
    router.push(`/team/${teamId}/stories`);
  };

  if (statsLoading) {
    return (
      <SafeContainer>
        <Header />
        <ActivityIndicator size="large" className="flex-1 justify-center" />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer>
      <Header />
      <ScrollView className="flex-1">
        <Overview stats={stats} />
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
