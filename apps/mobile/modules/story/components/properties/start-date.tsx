import type { Story } from "@/modules/stories/types";
import { DateField } from "@/components/ui/date-field";

export function StartDateBadge({
  story,
  disabled,
  onStartDateChange,
}: {
  story: Story;
  disabled?: boolean;
  onStartDateChange: (date: Date | null) => Promise<void>;
}) {
  return (
    <DateField
      disabled={disabled}
      label="Start date"
      value={story.startDate}
      onChange={onStartDateChange}
    />
  );
}
