import type { StoryActivity } from "@/modules/stories/types";

const ASSOCIATION_REASONS = new Set([
  "association_added",
  "association_updated",
  "association_removed",
]);

const ASSOCIATION_FIELDS = new Set([
  "blocked_by_id",
  "blocking_id",
  "related_id",
  "duplicate_id",
  "duplicated_by_id",
]);

type ActivityCopySegment =
  | { type: "currentValue" }
  | { type: "fieldLabel" }
  | { type: "oldValue"; value: string }
  | { text: string; type: "text" };

type ActivityCopy = {
  segments: ActivityCopySegment[];
  text: string;
};

type ActivityCopyInput = {
  currentValue: string;
  field: string;
  fieldLabel: string;
  oldValue?: unknown;
  reason?: StoryActivity["reason"];
  type: StoryActivity["type"];
};

export const getDisplayActivityReason = (reason?: string | null) => {
  const normalizedReason = reason?.trim() ?? "";
  return ASSOCIATION_REASONS.has(normalizedReason) ? "" : normalizedReason;
};

export const getActivityCopy = ({
  currentValue,
  field,
  fieldLabel,
  oldValue,
  reason,
  type,
}: ActivityCopyInput): ActivityCopy => {
  if (type === "create") {
    return buildCopy([{ text: "created the story", type: "text" }], {
      currentValue,
      fieldLabel,
    });
  }

  if (type === "link") {
    return buildCopy([{ text: "linked", type: "text" }], {
      currentValue,
      fieldLabel,
    });
  }

  const associationCopy = getAssociationActivityCopy({
    currentValue,
    field,
    fieldLabel,
    oldValue,
    reason,
  });
  if (associationCopy) {
    return associationCopy;
  }

  const oldValueText = stringifyActivityValue(oldValue);
  const segments = getFieldUpdateSegments(field, oldValueText);
  return buildCopy(segments, { currentValue, fieldLabel });
};

const getAssociationActivityCopy = ({
  currentValue,
  field,
  fieldLabel,
  oldValue,
  reason,
}: Omit<ActivityCopyInput, "type">): ActivityCopy | null => {
  const normalizedReason = reason?.trim() ?? "";
  if (!ASSOCIATION_FIELDS.has(field) && !ASSOCIATION_REASONS.has(normalizedReason)) {
    return null;
  }

  if (normalizedReason === "association_removed") {
    return buildCopy(
      [
        { text: "removed the", type: "text" },
        { type: "fieldLabel" },
        { text: "relationship with", type: "text" },
        { type: "currentValue" },
      ],
      { currentValue, fieldLabel },
    );
  }

  const oldValueText = stringifyActivityValue(oldValue);
  if (normalizedReason === "association_updated" && oldValueText) {
    const oldRelationshipLabel = normalizeRelationshipLabelForSentence(oldValueText);
    return buildCopy(
      [
        { text: "changed", type: "text" },
        { type: "currentValue" },
        { text: "from", type: "text" },
        { type: "oldValue", value: oldRelationshipLabel },
        { text: "to", type: "text" },
        { type: "fieldLabel" },
      ],
      { currentValue, fieldLabel },
    );
  }

  return buildCopy(
    [
      { text: "marked", type: "text" },
      { type: "currentValue" },
      { text: "as", type: "text" },
      { type: "fieldLabel" },
    ],
    { currentValue, fieldLabel },
  );
};

const getFieldUpdateSegments = (
  field: string,
  oldValueText: string,
): ActivityCopySegment[] => {
  switch (field) {
    case "title":
      return [
        { text: "renamed the story to", type: "text" },
        { type: "currentValue" },
      ];
    case "description":
    case "description_html":
      return [{ text: "updated the description", type: "text" }];
    case "status_id":
      return withOptionalOldValue(oldValueText, "moved the story");
    case "assignee_id":
      return oldValueText
        ? withOptionalOldValue(oldValueText, "reassigned the story")
        : [
            { text: "assigned the story to", type: "text" },
            { type: "currentValue" },
          ];
    case "priority":
      return withOptionalOldValue(oldValueText, "changed priority");
    case "estimate_unit":
      return [
        { text: "estimated the story at", type: "text" },
        { type: "currentValue" },
      ];
    case "sprint_id":
      return withOptionalOldValue(oldValueText, "moved the story");
    case "objective_id":
      return withOptionalOldValue(oldValueText, "moved the story");
    case "start_date":
      return [
        { text: "set the start date to", type: "text" },
        { type: "currentValue" },
      ];
    case "end_date":
      return [
        { text: "set the deadline to", type: "text" },
        { type: "currentValue" },
      ];
    case "labels":
      return [{ text: "updated labels", type: "text" }];
    default:
      return [
        { text: "changed", type: "text" },
        { type: "fieldLabel" },
        { text: "to", type: "text" },
        { type: "currentValue" },
      ];
  }
};

const withOptionalOldValue = (
  oldValueText: string,
  prefix: string,
): ActivityCopySegment[] => {
  if (!oldValueText) {
    return [
      { text: `${prefix} to`, type: "text" },
      { type: "currentValue" },
    ];
  }

  return [
    { text: `${prefix} from`, type: "text" },
    { type: "oldValue", value: oldValueText },
    { text: "to", type: "text" },
    { type: "currentValue" },
  ];
};

const buildCopy = (
  segments: ActivityCopySegment[],
  values: { currentValue: string; fieldLabel: string },
): ActivityCopy => ({
  segments,
  text: segments
    .map((segment) => {
      if (segment.type === "text") return segment.text;
      if (segment.type === "fieldLabel") return values.fieldLabel;
      if (segment.type === "oldValue") return segment.value;
      return values.currentValue;
    })
    .filter(Boolean)
    .join(" "),
});

const stringifyActivityValue = (value: unknown): string => {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return "";
};

const normalizeRelationshipLabelForSentence = (value: string) => {
  if (value === "Related to") return "Related";
  return value;
};
