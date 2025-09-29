import React from "react";
import { ScrollView } from "react-native";
import { Header } from "./components/header";
import { WorkItemList } from "./components/work-item-list";

export const MyWork = () => {
  return (
    <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
      <Header />
      <WorkItemList />
    </ScrollView>
  );
};
