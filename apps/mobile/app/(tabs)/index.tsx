import React from "react";
import { View, StyleSheet, ScrollView } from "react-native";
import { Header } from "../../components/shared/Header";
import { Section } from "../../components/shared/Section";
import { TeamLink } from "../../components/shared/TeamLink";
import { StatsCard } from "../../components/shared/StatsCard";

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
    <View style={styles.container}>
      <Header title="Hello, Joseph" onSettingsPress={handleMenuPress} />

      <ScrollView
        style={styles.scrollView}
        showsVerticalScrollIndicator={false}
      >
        <Section title="Overview">
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
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#FFFFFF",
  },
  scrollView: {
    flex: 1,
  },
  statsContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    paddingHorizontal: 12,
    paddingVertical: 8,
    gap: 6,
  },
});
