import React from "react";
import { Header } from "./components/header";
import { WorkItemList } from "./components/work-item-list";
import { SafeContainer } from "@/components/ui";

export const MyWork = () => {
  return (
    <SafeContainer isFull>
      <Header />
      <WorkItemList />
    </SafeContainer>
  );
};
