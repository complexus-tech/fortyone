import type { Story } from "@/modules/stories/types";
import { DateField } from "@/components/ui/date-field";

export function EndDateBadge({
  story,
  disabled,
  onEndDateChange,
}: {
  story: Story;
  disabled?: boolean;
  onEndDateChange: (date: Date | null) => Promise<void>;
}) {
  return (
    <DateField
      disabled={disabled}
      label="Deadline"
      value={story.endDate}
      onChange={onEndDateChange}
    />
  );
}
