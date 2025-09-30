import React from "react";
import { Header } from "./components/header";
import { StoriesList } from "./components/list";
import { SafeContainer, Tabs } from "@/components/ui";

export const MyWork = () => {
  return (
    <SafeContainer isFull>
      <Header />
      <Tabs defaultValue="all">
        <Tabs.List>
          <Tabs.Tab value="all">All</Tabs.Tab>
          <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
          <Tabs.Tab value="created">Created</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <StoriesList />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <StoriesList />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesList />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
