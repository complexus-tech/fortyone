import type { StatusCategory } from "@/types/statuses";
import type { FeedbackItem } from "./types";
import { StatusIcon } from "@/components/icons";
import { colors } from "@/constants/colors";
const META: Record<
  FeedbackItem["status"],
  { category: StatusCategory; color?: string }
> = {
  pending: { category: "backlog" },
  reviewing: { category: "started", color: colors.info },
  planned: { category: "unstarted", color: colors.primary },
  in_progress: { category: "started", color: colors.warning },
  completed: { category: "completed", color: colors.success },
  closed: { category: "cancelled", color: colors.danger },
};
export function FeedbackStatusIcon({
  status,
  size = 20,
}: {
  status: FeedbackItem["status"];
  size?: number;
}) {
  return <StatusIcon {...META[status]} size={size} />;
}
