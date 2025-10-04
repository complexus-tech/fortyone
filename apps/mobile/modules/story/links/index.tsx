import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { SafeContainer, StoriesSkeleton } from "@/components/ui";
import { useLinks } from "./hooks/use-links";
import { List } from "./components";

export const Links = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: links = [], isPending } = useLinks(storyId!);

  if (isPending) {
    return (
      <SafeContainer>
        <StoriesSkeleton count={3} />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer>
      <List links={links} />
    </SafeContainer>
  );
};
