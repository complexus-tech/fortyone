"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { TeamOptionsProvider } from "./provider";
import { Header } from "./header";
import { AllStories } from "./all-stories";

export const ListBacklogStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:backlog:layout",
    "list",
  );

  return (
    <TeamOptionsProvider>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} />
    </TeamOptionsProvider>
  );
};
