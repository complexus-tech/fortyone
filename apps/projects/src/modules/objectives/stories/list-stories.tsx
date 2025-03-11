"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { ObjectiveOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Header } from "./header";

export const ListStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "kanban",
  );

  return (
    <ObjectiveOptionsProvider>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} />
    </ObjectiveOptionsProvider>
  );
};
