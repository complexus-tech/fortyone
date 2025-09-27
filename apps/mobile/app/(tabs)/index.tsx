import React from "react";
import { View, StyleSheet, ScrollView } from "react-native";
import { Header } from "../../components/shared/Header";
import { Section } from "../../components/shared/Section";
import { TeamLink } from "../../components/shared/TeamLink";

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
      <Header title="Home" onSettingsPress={handleMenuPress} />

      <ScrollView
        style={styles.scrollView}
        showsVerticalScrollIndicator={false}
      >
        <Section title="Teams">
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
});
