import React from "react";
import { Text } from "@/components/ui";

type StoryGroupHeaderProps = {
  title: string;
};

export const StoryGroupHeader = ({ title }: StoryGroupHeaderProps) => {
  return (
    <Text fontWeight="semibold" color="muted" className="pb-2 pt-3 px-4">
      {title}
    </Text>
  );
};
