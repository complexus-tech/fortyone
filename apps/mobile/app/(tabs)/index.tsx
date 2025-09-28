import React from "react";
import { View, StyleSheet, ScrollView } from "react-native";
import { Header } from "../../components/shared/Header";
import { Section } from "../../components/shared/Section";
import { TeamLink } from "../../components/shared/TeamLink";
import { StatsCard } from "../../components/shared/StatsCard";
import { Container, Text } from "@/components/ui";

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
        <Section title="Overview">
          <Text style={styles.statsText}>
            Here&apos;s what&apos;s happening with your stories.
          </Text>
          <View style={styles.statsContainer}>
            <StatsCard
              title="Stories Created"
              count={120}
              onPress={() => console.log("Stories created pressed")}
            />
            <StatsCard
              title="Stories closed"
              count={166}
              onPress={() => console.log("Stories closed pressed")}
            />
          </View>
        </Section>

        <Section title="YourTeams">
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

        <Text>Edge-to-edge content</Text>

        <View className="bg-gray-50 py-4">
          <Text color="success">
            Container with background and vertical padding
          </Text>
        </View>
      </ScrollView>
    </Container>
  );
}

const styles = StyleSheet.create({
  statsContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    gap: 6,
  },
  statsText: {
    fontSize: 15,
    fontWeight: "400",
    color: "#666",
    marginBottom: 2,
    marginTop: 4,
  },
});
