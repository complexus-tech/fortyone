"use client";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/types/story";
import { useMilestoneStories } from "./provider";

export const AllStories = ({
  layout,
  stories,
}: {
  layout: StoriesLayout;
  stories: Story[];
}) => {
  const { viewOptions } = useMilestoneStories();
  return (
    <StoriesBoard layout={layout} stories={stories} viewOptions={viewOptions} />
  );
};
