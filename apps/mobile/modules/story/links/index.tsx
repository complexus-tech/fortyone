import React from "react";
import { useLocalSearchParams } from "expo-router";
import { StoriesSkeleton } from "@/components/ui";
import { useLinks } from "./hooks/use-links";
import { List } from "./components";
import { QueryState } from "@/components/ui/query-state";

export const Links = () => {
  const { storyId } = useLocalSearchParams<{ storyId: string }>();
  const { data: links, isPending, error, refetch } = useLinks(storyId);

  if (isPending) {
    return <StoriesSkeleton count={3} />;
  }

  if (error && !links)
    return (
      <QueryState
        title="Couldn’t load links"
        message="Check your connection and try again."
        onRetry={() => {
          void refetch();
        }}
      />
    );

  return <List links={links ?? []} />;
};
