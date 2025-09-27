import React from "react";
import { View, StyleSheet, ScrollView } from "react-native";
import { Header } from "../../components/shared/Header";
import { Section } from "../../components/shared/Section";
import { CycleItem } from "../../components/shared/CycleItem";

export default function MyWork() {
  const handleMenuPress = () => {
    console.log("Menu pressed");
    // Show menu options
  };

  const handleCyclePress = (cycleTitle: string) => {
    console.log("Cycle pressed:", cycleTitle);
    // Navigate to cycle details
  };

  // Mock data for work items
  const workItems = [
    {
      title: "hey",
      status: "in-review" as const,
      assignee: { name: "John Doe" },
    },
    {
      title: "check again",
      status: "in-progress" as const,
      assignee: { name: "Jane Smith" },
    },
    { title: "upgrade nextjs", status: "todo" as const },
  ];

  return (
    <View style={styles.container}>
      <Header title="My Work" onSettingsPress={handleMenuPress} />

      <ScrollView
        style={styles.scrollView}
        showsVerticalScrollIndicator={false}
      >
        <Section title="Work Items">
          {workItems.map((item, index) => (
            <CycleItem
              key={index}
              title={item.title}
              status={item.status}
              assignee={item.assignee}
              onPress={() => handleCyclePress(item.title)}
            />
          ))}
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
