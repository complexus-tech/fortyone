"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { TeamOptionsProvider } from "./provider";
import { Header } from "./header";
import { AllStories } from "./all-stories";

export const ListArchivedStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:archived:layout",
    "list",
  );

  return (
    <TeamOptionsProvider layout={layout}>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} />
    </TeamOptionsProvider>
  );
};
