import React from "react";
import { ScrollView } from "react-native";
import { Header } from "./components/header";
import { Container, Section } from "@/components/ui";
import { Team } from "./components/team";
import { SafeAreaView } from "react-native-safe-area-context";

export const Home = () => {
  return (
    <Container>
      <SafeAreaView>
        <Header />
        <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
          <Section title="Your Teams">
            <Team name="Engineering" color="#FF9500" />
            <Team name="Complexus" color="#34C759" />
          </Section>
        </ScrollView>
      </SafeAreaView>
    </Container>
  );
};
