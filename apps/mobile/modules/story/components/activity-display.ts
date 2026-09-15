import type { StoryActivity, StoryActivityUser } from "@/modules/stories/types";
import { format, formatDistance, isValid, parseISO } from "date-fns";

type Person = Pick<StoryActivityUser, "id" | "username" | "fullName"> & {
  isSystem?: boolean;
  role?: string;
};
type NamedReference = { id: string; name: string };

export type ActivityDisplayContext = {
  members: ReadonlyMap<string, Person>;
  statuses: ReadonlyMap<string, NamedReference>;
  sprints: ReadonlyMap<string, NamedReference>;
  objectives: ReadonlyMap<string, NamedReference>;
  labels: ReadonlyMap<string, NamedReference>;
  currentUserId?: string | null;
  timezone?: string;
  storyTerm: string;
  sprintTerm: string;
  objectiveTerm: string;
  keyResultTerm: string;
  estimateScheme: "points" | "tshirt";
};

export const indexActivityReferences = <T extends { id: string }>(
  items: readonly T[],
): ReadonlyMap<string, T> => new Map(items.map((item) => [item.id, item]));

export function indexActivityPeople(
  members: readonly Person[],
  mayaAssignee?: Person | null,
): ReadonlyMap<string, Person> {
  const people = new Map(members.map((member) => [member.id, member]));
  // /members excludes system accounts. Use the separate API identity for
  // assignment values, never a guessed ID or a name inferred from the field.
  if (mayaAssignee?.id && mayaAssignee.isSystem) {
    people.set(mayaAssignee.id, mayaAssignee);
  }
  return people;
}

const UUID = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i;
const EMPTY_VALUES = new Set([
  "",
  "nil",
  "<nil>",
  "null",
  "undefined",
  "00000000-0000-0000-0000-000000000000",
]);
const ASSOCIATIONS: Record<string, string> = {
  blocked_by_id: "blocked by",
  blocking_id: "blocking",
  related_id: "related to",
  duplicate_id: "duplicate of",
  duplicated_by_id: "duplicated by",
};
const FIELD_LABELS: Record<string, string> = {
  status_id: "status",
  assignee_id: "assignee",
  reporter_id: "reporter",
  estimate_unit: "complexity",
  estimated_duration_minutes: "time needed",
  minimum_focus_block_minutes: "minimum focus block",
  start_date: "start date",
  end_date: "deadline",
  parent_id: "parent task",
  epic_id: "epic",
  collaborator_ids: "collaborators",
  auto_scheduling_status: "auto-scheduling",
  auto_scheduling_time: "scheduled time",
  auto_scheduling_locked: "schedule lock",
  auto_scheduling_enabled: "auto-scheduling",
};
const TSHIRT_ESTIMATES: Record<number, string> = {
  1: "XS",
  2: "S",
  3: "M",
  5: "L",
  8: "XL",
};

const scalar = (value: unknown): string =>
  typeof value === "string"
    ? value.trim()
    : typeof value === "number" || typeof value === "boolean"
      ? String(value)
      : "";
const isMissing = (value: unknown) =>
  EMPTY_VALUES.has(scalar(value).toLowerCase());
const humanize = (value: string) => value.replaceAll("_", " ");
const sentenceCase = (value: string) =>
  value.replace(/^./, (first) => first.toUpperCase());
const personName = (person?: Person | null) =>
  person?.fullName?.trim() || person?.username?.trim();

export function resolveActivityActor(
  activity: Pick<StoryActivity, "userId" | "user">,
  context: Pick<ActivityDisplayContext, "members" | "currentUserId">,
) {
  // The embedded account survives workspace membership changes and includes
  // system actors. Never infer an actor from the kind of change they made.
  const person: Person | undefined | null =
    activity.user?.id === activity.userId
      ? activity.user
      : context.members.get(activity.userId);
  const isSystem = person?.isSystem === true || person?.role === "system";
  const isSelf =
    Boolean(activity.userId) && activity.userId === context.currentUserId;
  return {
    name: isSelf
      ? "You"
      : personName(person) || (isSystem ? "System" : "Someone"),
    isSystem,
  };
}

const scheduleFormatters = new Map<string, Intl.DateTimeFormat>();

function timestampText(value: unknown, timezone?: string): string | null {
  if (typeof value !== "string" || !timezone?.trim()) return null;
  const date = parseISO(value);
  if (!isValid(date)) return null;
  try {
    const zone = timezone.trim();
    let formatter = scheduleFormatters.get(zone);
    if (!formatter) {
      formatter = new Intl.DateTimeFormat("en-GB", {
        day: "numeric",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        hourCycle: "h23",
        timeZone: zone,
      });
      scheduleFormatters.set(zone, formatter);
    }
    const parts = new Map(
      formatter.formatToParts(date).map(({ type, value }) => [type, value]),
    );
    return `${parts.get("day")} ${parts.get("month")} ${parts.get("year")} at ${parts.get("hour")}:${parts.get("minute")}`;
  } catch {
    // A stale/invalid profile timezone must not break the activity feed.
    return null;
  }
}

export function formatScheduleActivityValue(
  currentValue: string,
  newValue: unknown,
  timezone?: string,
) {
  const formatted = timestampText(newValue, timezone);
  if (formatted) return formatted;
  // Legacy rows already contain the calendar owner's formatted local time.
  // Match web display without reparsing that string in the device timezone.
  const legacy = currentValue
    .trim()
    .replace(/\s+(?:[A-Z]{2,6}|[+-]\d{2,4})$/, "");
  return !isMissing(legacy) ? legacy : "an unavailable time";
}

export function formatActivityTimestamp(createdAt: string, now = new Date()) {
  const date = parseISO(createdAt);
  return isValid(date) && isValid(now)
    ? formatDistance(date, now, { addSuffix: true })
    : null;
}

function dateOnlyText(value: unknown) {
  const raw = scalar(value).split(/[T ]/)[0];
  if (!/^\d{4}-\d{2}-\d{2}$/.test(raw)) return "an unavailable date";
  const date = parseISO(raw);
  return isValid(date) ? format(date, "d MMM yyyy") : "an unavailable date";
}

function durationText(value: unknown, fieldLabel: string) {
  const minutes = Number(scalar(value));
  if (!Number.isFinite(minutes) || minutes <= 0 || !Number.isInteger(minutes)) {
    return `no ${fieldLabel}`;
  }
  const hours = Math.floor(minutes / 60);
  const remainder = minutes % 60;
  return [
    hours ? `${hours} ${hours === 1 ? "hour" : "hours"}` : "",
    remainder ? `${remainder} ${remainder === 1 ? "minute" : "minutes"}` : "",
  ]
    .filter(Boolean)
    .join(" ");
}

// Modern values are JSON arrays; pre-migration rows used Go's [uuid uuid] format.
function referenceIds(value: unknown): string[] {
  if (Array.isArray(value))
    return value.filter(
      (item): item is string => typeof item === "string" && UUID.test(item),
    );
  if (typeof value !== "string") return [];
  return value.match(/[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}/gi) ?? [];
}

function referenceValue(
  value: unknown,
  items: ReadonlyMap<string, NamedReference>,
  label: string,
) {
  const raw = scalar(value);
  if (isMissing(value)) return `no ${label}`;
  return (
    items.get(raw)?.name || (UUID.test(raw) ? `an unavailable ${label}` : raw)
  );
}

function fieldValue(
  field: string,
  value: unknown,
  context: ActivityDisplayContext,
  activity: StoryActivity,
): string {
  const raw = scalar(value);
  switch (field) {
    case "title":
      return raw;
    case "assignee_id":
    case "reporter_id": {
      if (isMissing(value)) return "Unassigned";
      const person =
        activity.user?.id === raw ? activity.user : context.members.get(raw);
      return (
        personName(person) || (UUID.test(raw) ? "an unavailable member" : raw)
      );
    }
    case "status_id":
      return referenceValue(value, context.statuses, "status");
    case "sprint_id":
      return referenceValue(value, context.sprints, context.sprintTerm);
    case "objective_id":
      return referenceValue(value, context.objectives, context.objectiveTerm);
    case "auto_scheduling_status":
      return sentenceCase(humanize(raw));
    case "start_date":
    case "end_date":
      return isMissing(value)
        ? `no ${FIELD_LABELS[field]}`
        : dateOnlyText(value);
    case "estimated_duration_minutes":
    case "minimum_focus_block_minutes":
      return durationText(value, FIELD_LABELS[field]);
    case "estimate_unit": {
      if (isMissing(value) || raw === "0") return "no complexity";
      const estimate = Number(raw);
      if (!Number.isInteger(estimate) || estimate < 0) return raw;
      if (context.estimateScheme === "tshirt") {
        return TSHIRT_ESTIMATES[estimate] || raw;
      }
      return `${estimate} ${estimate === 1 ? "point" : "points"}`;
    }
    default:
      return isMissing(value)
        ? ""
        : UUID.test(raw)
          ? "an unavailable value"
          : raw;
  }
}

function collectionValue(
  field: "labels" | "collaborator_ids",
  activity: StoryActivity,
  context: ActivityDisplayContext,
) {
  const raw = activity.newValue ?? activity.currentValue;
  const ids = referenceIds(raw);
  const label = field === "labels" ? "label" : "collaborator";
  if (ids.length) {
    return ids
      .map((id) =>
        field === "labels"
          ? context.labels.get(id)?.name || "an unavailable label"
          : personName(context.members.get(id)) || "an unavailable member",
      )
      .join(", ");
  }
  if (Array.isArray(raw) || isMissing(raw) || scalar(raw) === "[]")
    return `no ${label}s`;
  return scalar(raw);
}

function activityMessage(
  activity: StoryActivity,
  context: ActivityDisplayContext,
) {
  const { field, type, currentValue, oldValue, reason } = activity;
  const term = context.storyTerm;
  if (type === "create") return `created the ${term}`;
  if (type === "link")
    return currentValue ? `linked ${currentValue}` : "added a link";
  if (field === "description" || field === "description_html")
    return "updated the description";
  const current = fieldValue(field, currentValue, context, activity);
  const previous = isMissing(oldValue)
    ? ""
    : fieldValue(field, oldValue, context, activity);
  const move = (verb: string) =>
    `${verb}${previous ? ` from ${previous}` : ""} to ${current}`;
  const relation = ASSOCIATIONS[field];
  if (relation) {
    if (reason === "association_removed")
      return `removed the ${relation} relationship with ${current}`;
    if (reason === "association_updated" && previous)
      return `changed ${current} from ${previous.toLowerCase()} to ${relation}`;
    return `marked ${current} as ${relation}`;
  }
  switch (field) {
    case "title":
      return `renamed the ${term} to ${current}`;
    case "status_id":
      return move(`moved the ${term}`);
    case "priority":
      return move("changed priority");
    case "assignee_id":
      return isMissing(currentValue)
        ? `unassigned the ${term}`
        : move(`${previous ? "reassigned" : "assigned"} the ${term}`);
    case "sprint_id":
    case "objective_id":
      return isMissing(currentValue)
        ? `removed the ${term} from ${previous || `its ${field === "sprint_id" ? context.sprintTerm : context.objectiveTerm}`}`
        : move(`moved the ${term}`);
    case "start_date":
    case "end_date":
      return isMissing(currentValue)
        ? `removed the ${FIELD_LABELS[field]}`
        : `set the ${FIELD_LABELS[field]} to ${current}`;
    case "estimate_unit":
    case "estimated_duration_minutes":
    case "minimum_focus_block_minutes":
      return `set the ${term} ${FIELD_LABELS[field]} to ${current}`;
    case "auto_scheduling_time":
      return `${isMissing(oldValue) ? "scheduled work for" : "rescheduled work to"} ${formatScheduleActivityValue(currentValue, activity.newValue, context.timezone)}`;
    case "auto_scheduling_status":
      return previous.toLowerCase() === current.toLowerCase()
        ? `updated auto-scheduling status to ${current}`
        : move("changed auto-scheduling");
    case "auto_scheduling_locked":
    case "auto_scheduling_enabled": {
      const value = scalar(activity.newValue ?? currentValue).toLowerCase();
      if (value !== "true" && value !== "false")
        return "updated auto-scheduling";
      return field === "auto_scheduling_enabled"
        ? `${value === "true" ? "enabled" : "paused"} auto-scheduling`
        : `${value === "true" ? "locked" : "unlocked"} the auto-scheduled calendar blocks`;
    }
    case "labels":
    case "collaborator_ids":
      return `updated ${field === "labels" ? "labels" : "collaborators"} to ${collectionValue(field, activity, context)}`;
    default: {
      const label =
        field === "key_result_id"
          ? context.keyResultTerm
          : FIELD_LABELS[field] || humanize(field).replace(/ id$/, "");
      return current
        ? `changed ${label || "the details"} to ${current}`
        : `cleared ${label || "the value"}`;
    }
  }
}

export function getActivityDisplay(
  activity: StoryActivity,
  context: ActivityDisplayContext,
) {
  const actor = resolveActivityActor(activity, context);
  // Mobile omits explanatory rationale. Association reason codes still inform
  // the event verb inside activityMessage, so removals retain their meaning.
  return { actor, message: activityMessage(activity, context) };
}
