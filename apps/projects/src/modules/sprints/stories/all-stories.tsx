"use client";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { useSprintStories } from "@/modules/stories/hooks/sprint-stories";
import { useSprintOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { sprintId } = useParams<{ sprintId: string }>();
  const { viewOptions } = useSprintOptions();
  const { data: stories = [] } = useSprintStories(sprintId);
  return (
    <StoriesBoard layout={layout} stories={stories} viewOptions={viewOptions} />
  );
};
