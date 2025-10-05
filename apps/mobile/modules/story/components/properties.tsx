import { useStory } from "@/modules/stories/hooks";
import { useStatuses } from "@/modules/statuses/hooks/use-statuses";
import React from "react";
import { useGlobalSearchParams } from "expo-router";
import { Row, Badge, Text } from "@/components/ui";
import { hexToRgba } from "@/lib/utils/colors";
import { Dot } from "@/components/icons";

export const Properties = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  const { data: statuses = [] } = useStatuses();

  const status = statuses.find((s) => s.id === story?.statusId);

  return (
    <Row wrap gap={2} asContainer className="my-4">
      <Badge
        rounded="xl"
        style={{
          backgroundColor: hexToRgba(status?.color || "#6B665C", 0.1),
          borderColor: hexToRgba(status?.color || "#6B665C", 0.2),
          borderWidth: 1,
        }}
      >
        <Row align="center" gap={1}>
          <Dot color={status?.color} size={12} />
          <Text>{status?.name || "No Status"}</Text>
        </Row>
      </Badge>
    </Row>
  );
};
