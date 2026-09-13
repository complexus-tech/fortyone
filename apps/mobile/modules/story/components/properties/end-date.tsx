import type { Story } from "@/modules/stories/types";
import { DateField } from "@/components/ui/date-field";

export function EndDateBadge({
  story,
  onEndDateChange,
}: {
  story: Story;
  onEndDateChange: (date: Date | null) => void;
}) {
  return (
    <DateField
      label="Deadline"
      value={story.endDate}
      onChange={onEndDateChange}
    />
  );
}
