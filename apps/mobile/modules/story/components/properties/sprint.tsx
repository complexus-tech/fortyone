import type { Story } from "@/modules/stories/types";
import { useColorScheme } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTerminology } from "@/hooks";
import { useTeamSprints } from "@/modules/sprints/hooks";
import { PropertyChip } from "./property-chip";
import { PropertyBottomSheet } from "./property-bottom-sheet";

export const SprintBadge = ({
  story,
  disabled,
  onSprintChange,
}: {
  story: Story;
  disabled?: boolean;
  onSprintChange: (id: string | null) => Promise<void>;
}) => {
  const { getTermDisplay } = useTerminology();
  const dark = useColorScheme() === "dark";
  const {
    data: items = [],
    isPending,
    error,
    refetch,
  } = useTeamSprints(story.teamId);
  const current = items.find((item) => item.id === story.sprintId);
  const title = getTermDisplay("sprintTerm", { capitalize: true });
  const icon = (
    <Ionicons
      name="play-circle-outline"
      size={16}
      color={themeColors[dark ? "dark" : "light"].textMuted}
    />
  );
  return (
    <PropertyBottomSheet
      title={title}
      disabled={disabled}
      loading={isPending}
      error={error}
      onRetry={() => {
        void refetch();
      }}
      trigger={
        <PropertyChip>
          {icon}
          <Text
            fontSize="sm"
            numberOfLines={1}
            style={{ maxWidth: 150, flexShrink: 1 }}
          >
            {current?.name || `Add ${title.toLowerCase()}`}
          </Text>
        </PropertyChip>
      }
      options={items.map((item) => ({ id: item.id, label: item.name, icon }))}
      selectedIds={story.sprintId ? [story.sprintId] : []}
      onSelect={onSprintChange}
      clearLabel={`No ${title.toLowerCase()}`}
      onClear={() => onSprintChange(null)}
    />
  );
};
