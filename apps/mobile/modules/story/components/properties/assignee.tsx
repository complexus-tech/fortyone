import type { Story } from "@/modules/stories/types";
import { Text, Avatar } from "@/components/ui";
import { useMembers } from "@/modules/members/hooks/use-members";
import { PropertyChip } from "./property-chip";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { memberDisplayName } from "./picker-utils";

export const AssigneeBadge = ({
  story,
  disabled,
  onAssigneeChange,
  label = "Assignee",
}: {
  story: Pick<Story, "assigneeId">;
  disabled?: boolean;
  label?: string;
  onAssigneeChange: (id: string | null) => Promise<void>;
}) => {
  const { data: members = [], isPending, error, refetch } = useMembers();
  const current = members.find((member) => member.id === story.assigneeId);
  const name = current ? memberDisplayName(current) : "Unassigned";
  return (
    <PropertyBottomSheet
      title={label}
      disabled={disabled}
      loading={isPending}
      error={error}
      onRetry={() => {
        void refetch();
      }}
      trigger={
        <PropertyChip>
          <Avatar
            size="xs"
            name={current ? name : undefined}
            src={current?.avatarUrl}
          />
          <Text
            fontSize="sm"
            numberOfLines={1}
            style={{ maxWidth: 150, flexShrink: 1 }}
          >
            {name}
          </Text>
        </PropertyChip>
      }
      options={members
        .filter(
          (member) =>
            member.role !== "system" &&
            (member.isActive || member.id === story.assigneeId),
        )
        .map((member) => ({
          id: member.id,
          label: memberDisplayName(member),
          description: member.email,
          icon: (
            <Avatar
              size="sm"
              name={memberDisplayName(member)}
              src={member.avatarUrl}
            />
          ),
        }))}
      selectedIds={story.assigneeId ? [story.assigneeId] : []}
      onSelect={onAssigneeChange}
      clearLabel="Unassigned"
      onClear={() => onAssigneeChange(null)}
    />
  );
};
