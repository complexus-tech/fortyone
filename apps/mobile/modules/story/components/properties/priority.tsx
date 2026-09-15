import type { StoryPriority } from "@/modules/stories/types";
import { Text } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { PropertyChip } from "./property-chip";
import { PropertyBottomSheet } from "./property-bottom-sheet";
const PRIORITIES: StoryPriority[] = [
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
];
export const PriorityBadge = ({
  priority,
  disabled,
  onPriorityChange,
}: {
  priority: StoryPriority;
  disabled?: boolean;
  onPriorityChange: (priority: StoryPriority) => Promise<void>;
}) => (
  <PropertyBottomSheet
    title="Priority"
    searchable={false}
    disabled={disabled}
    trigger={
      <PropertyChip>
        <PriorityIcon priority={priority} />
        <Text fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
          {priority}
        </Text>
      </PropertyChip>
    }
    options={PRIORITIES.map((value) => ({
      id: value,
      label: value,
      icon: <PriorityIcon size={20} priority={value} />,
    }))}
    selectedIds={[priority]}
    onSelect={(value) => onPriorityChange(value as StoryPriority)}
  />
);
