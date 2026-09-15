import type { Story } from "@/modules/stories/types";
import { Text } from "@/components/ui";
import { StatusIcon } from "@/components/icons";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { PropertyChip } from "./property-chip";
import { PropertyBottomSheet } from "./property-bottom-sheet";
export const StatusBadge = ({
  story,
  disabled,
  onStatusChange,
}: {
  story: Story;
  disabled?: boolean;
  onStatusChange: (id: string) => Promise<void>;
}) => {
  const {
    data: statuses = [],
    isPending,
    error,
    refetch,
  } = useTeamStatuses(story.teamId);
  const current = statuses.find((status) => status.id === story.statusId);
  return (
    <PropertyBottomSheet
      title="Status"
      searchable={false}
      disabled={disabled}
      loading={isPending}
      error={error}
      onRetry={() => {
        void refetch();
      }}
      trigger={
        <PropertyChip>
          <StatusIcon
            category={current?.category}
            color={current?.color}
            size={16}
          />
          <Text fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
            {current?.name || "No status"}
          </Text>
        </PropertyChip>
      }
      options={statuses.map((status) => ({
        id: status.id,
        label: status.name,
        icon: (
          <StatusIcon
            category={status.category}
            color={status.color}
            size={18}
          />
        ),
      }))}
      selectedIds={[story.statusId]}
      onSelect={onStatusChange}
    />
  );
};
