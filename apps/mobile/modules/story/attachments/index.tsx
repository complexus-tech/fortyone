import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { StoriesSkeleton } from "@/components/ui";
import { useStoryAttachments } from "./hooks/use-attachments";
import { List } from "./components";

export const Attachments = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: attachments = [], isPending } = useStoryAttachments(storyId);

  if (isPending) {
    return <StoriesSkeleton count={3} />;
  }

  return <List attachments={attachments} />;
};
