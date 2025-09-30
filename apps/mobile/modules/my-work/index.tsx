import React, { useState } from "react";
import { Header } from "./components/header";
import { StoriesList } from "./components/list";
import { SafeContainer, Tabs } from "@/components/ui";

export const MyWork = () => {
  const [tab, setTab] = useState("all");
  const tabs = [
    { value: "all", label: "All" },
    { value: "assigned", label: "Assigned" },
    { value: "created", label: "Created" },
  ];

  return (
    <SafeContainer isFull>
      <Header />
      <Tabs tabs={tabs} defaultValue={tab} onValueChange={setTab} />
      <StoriesList />
    </SafeContainer>
  );
};
