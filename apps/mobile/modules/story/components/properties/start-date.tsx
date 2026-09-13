import type { Story } from "@/modules/stories/types";
import { DateField } from "@/components/ui/date-field";

export function StartDateBadge({
  story,
  onStartDateChange,
}: {
  story: Story;
  onStartDateChange: (date: Date | null) => void;
}) {
  return (
    <DateField
      label="Start date"
      value={story.startDate}
      onChange={onStartDateChange}
    />
  );
}
