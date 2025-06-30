"use client";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story } from "@/modules/stories/types";
import { Header } from "./components/header";
import { AllStories } from "./components/all-stories";
import { ProfileProvider } from "./components/provider";

export const ListUserStories = ({ stories }: { stories: Story[] }) => {
  const { userId } = useParams<{
    userId: string;
  }>();

  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    `stories:${userId}:layout`,
    "list",
  );
  return (
    <ProfileProvider>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} stories={stories} />
    </ProfileProvider>
  );
};
