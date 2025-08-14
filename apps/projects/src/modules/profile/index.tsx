"use client";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story } from "@/modules/stories/types";
import { Header } from "./components/header";
import { AllStories } from "./components/all-stories";
import { ProfileProvider } from "./components/provider";

export const ListUserStories = ({ stories }: { stories: Story[] }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "profile:stories:layout",
    "kanban",
  );
  return (
    <ProfileProvider layout={layout}>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} stories={stories} />
    </ProfileProvider>
  );
};
