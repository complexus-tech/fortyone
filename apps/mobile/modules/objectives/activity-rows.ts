import type { DetailFeedRow } from "@/modules/entity-details/feed";
import type { ObjectiveActivity } from "./detail-data";
import { format } from "date-fns";

const FIELD_LABELS: Record<string, string> = {
  status_id: "status",
  lead_user_id: "lead",
  start_date: "start date",
  end_date: "deadline",
  is_private: "visibility",
  name: "title",
};

export function objectiveActivityRows({
  activities,
  tab,
  members,
  statuses,
  objectiveTerm,
  keyResultTerm,
}: {
  activities: ObjectiveActivity[];
  tab: string;
  members: { id: string; fullName: string }[];
  statuses: { id: string; name: string }[];
  objectiveTerm: string;
  keyResultTerm: string;
}): DetailFeedRow[] {
  return activities.flatMap<DetailFeedRow>((entry) => {
    const author =
      entry.user?.fullName ||
      entry.user?.username ||
      members.find((member) => member.id === entry.userId)?.fullName ||
      "Former member";
    if (tab === "comments")
      return entry.comment?.trim()
        ? [
            {
              id: entry.id,
              kind: "comment",
              author,
              avatar: entry.user?.avatarUrl,
              body: entry.comment,
              format: entry.field === "comment" ? "markdown" : "plain",
              createdAt: entry.createdAt,
            },
          ]
        : [];
    if (entry.field === "comment" || entry.field === "completed_at") return [];

    let value: string | undefined = entry.currentValue;
    if (!value || value.includes("nil")) value = undefined;
    else if (entry.field === "description") value = undefined;
    else if (entry.field === "status_id")
      value = statuses.find((status) => status.id === value)?.name;
    else if (entry.field === "lead_user_id" || entry.field === "lead")
      value = members.find((member) => member.id === value)?.fullName;
    else if (entry.field === "is_private")
      value = value === "true" ? "private" : "team visible";
    else if (["start_date", "end_date"].includes(entry.field)) {
      const date = new Date(value.split(" ")[0]);
      value = Number.isNaN(date.getTime())
        ? undefined
        : format(date, "d MMM yyyy");
    }
    const entity =
      entry.updateType === "key_result" ? keyResultTerm : objectiveTerm;
    const body =
      entry.type === "create"
        ? `created this ${entity}`
        : entry.type === "delete"
          ? `deleted a ${entity}`
          : `updated ${FIELD_LABELS[entry.field] ?? entry.field.replaceAll("_", " ")}${value ? ` to ${value}` : ""}`;
    return [
      {
        id: entry.id,
        kind: "update",
        author,
        body,
        createdAt: entry.createdAt,
      },
    ];
  });
}
