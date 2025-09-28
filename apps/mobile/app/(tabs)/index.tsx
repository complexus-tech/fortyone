import React from "react";
import { ScrollView } from "react-native";
import { Header } from "../../components/shared/Header";
import { Section } from "../../components/ui/Section";
import { TeamLink } from "../../components/shared/TeamLink";
import { Container } from "@/components/ui";

export default function Home() {
  const handleMenuPress = () => {
    console.log("Menu pressed");
    // Show menu options
  };

  const handleTeamPress = (teamName: string) => {
    console.log("Team pressed:", teamName);
    // Navigate to team details
  };

  return (
    <Container>
      <Header title="Hello, Joseph" onSettingsPress={handleMenuPress} />
      <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
        <Section title="Your Teams">
          <TeamLink
            name="Engineering"
            color="#FF9500"
            onPress={() => handleTeamPress("Engineering")}
          />
          <TeamLink
            name="Complexus"
            color="#34C759"
            onPress={() => handleTeamPress("Complexus")}
          />
        </Section>
      </ScrollView>
    </Container>
  );
}
