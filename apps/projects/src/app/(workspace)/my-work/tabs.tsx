"use client";
import { Tabs } from "ui";

export const MainTabs = () => {
  return (
    <Tabs defaultValue="assigned">
      <Tabs.List>
        <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
        <Tabs.Tab value="created">Created</Tabs.Tab>
        <Tabs.Tab value="subscribed">Subscribed</Tabs.Tab>
      </Tabs.List>
    </Tabs>
  );
};
