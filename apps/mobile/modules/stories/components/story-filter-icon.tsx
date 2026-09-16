import { ObjectiveIcon } from "@/components/icons/objective";
import type { StoryFilterOption } from "./story-filters.types";
import { View } from "react-native";
import { AssigneeIcon, PriorityIcon, StatusIcon } from "@/components/icons";
import { WebIcon } from "@/components/icons/web-icon";

/** Use the same glyphs and sizing as StoryRow on both sheet implementations. */
export function StoryFilterIcon({
  facet,
  option,
}: {
  facet: "status" | "priority" | "assignee" | "objective" | "sprint";
  option?: StoryFilterOption;
}) {
  return (
    <View
      pointerEvents="none"
      style={{
        width: 20,
        height: 20,
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      {facet === "sprint" ? (
        <WebIcon name="sprint" />
      ) : facet === "objective" ? (
        <ObjectiveIcon />
      ) : facet === "status" ? (
        <StatusIcon
          category={option?.statusCategory}
          color={option?.color}
          size={20}
        />
      ) : facet === "priority" ? (
        <PriorityIcon priority={option?.priority ?? "High"} size={18} />
      ) : (
        <AssigneeIcon size={20} />
      )}
    </View>
  );
}
