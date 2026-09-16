import type { StatusCategory } from "@/types/statuses";

export const record = (value: unknown): Record<string, unknown> | null =>
  value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;

export function humanize(value: string) {
  return value
    .replace(/^tool-/, "")
    .replace(/Tool$/, "")
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/_/g, " ")
    .replace(/^./, (letter) => letter.toUpperCase());
}

const PRIVATE_FIELDS =
  /^(confirmed|confirmationToken|idempotencyKey|token|apiKey|accessToken|sessionCookie)$/i;

/** Preserve every proposed field, while never presenting transport credentials. */
export function approvalDetails(
  input: unknown,
  names: ReadonlyMap<string, string>,
  prefix = "",
): { label: string; value: string }[] {
  const source = record(input);
  if (!source) return [];
  return Object.entries(source).flatMap(([key, value]) => {
    if (
      PRIVATE_FIELDS.test(key) ||
      value === undefined ||
      (key === "descriptionHTML" && source.description)
    )
      return [];
    const label = `${prefix}${humanize(key).replace(/ Ids?$/, "")}`;
    if (record(value)) return approvalDetails(value, names, `${label} · `);
    if (Array.isArray(value) && value.some((item) => record(item))) {
      return value.flatMap((item, index) =>
        approvalDetails(item, names, `${label} ${index + 1} · `),
      );
    }
    const format = (entry: unknown): string => {
      if (entry === null) return "None";
      if (typeof entry === "boolean") return entry ? "Yes" : "No";
      const text = String(entry);
      return names.get(text) ?? text;
    };
    return [
      {
        label,
        value: Array.isArray(value)
          ? value.map(format).join(", ") || "None"
          : format(value),
      },
    ];
  });
}

export type StoryResult = {
  id: string;
  title: string;
  reference?: string;
  statusId?: string;
  status?: { category: StatusCategory; name?: string; color?: string };
};

const STORY_ID = /^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i;
const STATUS_CATEGORIES = new Set<StatusCategory>([
  "backlog",
  "unstarted",
  "started",
  "paused",
  "completed",
  "cancelled",
]);
const text = (value: unknown) =>
  typeof value === "string" && value.trim() ? value.trim() : undefined;

function resultStory(item: unknown): StoryResult | null {
  const story = record(item);
  if (
    !story ||
    typeof story.id !== "string" ||
    !STORY_ID.test(story.id) ||
    !text(story.title)
  )
    return null;
  const status = record(story.status);
  const category = status?.category ?? story.statusCategory;
  const color = text(status?.color ?? story.statusColor);
  const teamCode = text(story.teamCode ?? record(story.team)?.code);
  const sequence = story.sequenceId;
  return {
    id: story.id,
    title: text(story.title)!,
    reference:
      text(story.reference) ??
      (teamCode &&
      typeof sequence === "number" &&
      Number.isSafeInteger(sequence) &&
      sequence > 0
        ? `${teamCode}-${sequence}`
        : undefined),
    statusId: text(story.statusId),
    status:
      typeof category === "string" &&
      STATUS_CATEGORIES.has(category as StatusCategory)
        ? {
            category: category as StatusCategory,
            name: text(status?.name ?? story.statusName),
            // Workspace status colors are hex values. Malformed historical tool
            // output must not reach native SVG's color parser.
            color:
              color && /^#(?:[\da-f]{3,4}|[\da-f]{6}|[\da-f]{8})$/i.test(color)
                ? color
                : undefined,
          }
        : undefined,
  };
}

/** Only canonical story envelopes are navigable; generic data/results may be other entities. */
export function resultStories(output: unknown): StoryResult[] {
  const result = record(output);
  if (!result) return [];
  const candidates = [result.story];
  if (Array.isArray(result.stories)) {
    for (const item of result.stories) {
      const group = record(item);
      // listTeamStories returns story groups; search and creation return a flat list.
      if (Array.isArray(group?.stories)) candidates.push(...group.stories);
      else candidates.push(item);
    }
  }
  const stories = new Map<string, StoryResult>();
  for (const candidate of candidates) {
    const story = resultStory(candidate);
    if (story && !stories.has(story.id)) stories.set(story.id, story);
  }
  return [...stories.values()];
}

export function messageLinkURL(value: string): string | null {
  try {
    const url = new URL(value);
    return ["https:", "http:"].includes(url.protocol) &&
      !url.username &&
      !url.password
      ? url.href
      : null;
  } catch {
    return null;
  }
}
