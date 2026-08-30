import { cn } from "lib";
import { Text } from "ui";
import type { ActivityCopySegment } from "./activity-copy";
import {
  ActivityAssociationBadge,
  ActivityLabelValue,
  isAssociationActivityField,
  type ActivityFieldMeta,
  type ActivityLabel,
} from "./activity-field-renderers";

type ActivityUpdateSegmentsProps = {
  activityLabels: ActivityLabel[];
  currentValue: string;
  field: string;
  fieldMeta: ActivityFieldMeta;
  segments: ActivityCopySegment[];
};

const getSegmentKey = (segment: ActivityCopySegment) => {
  if (segment.type === "text") return `text-${segment.text}`;
  if (segment.type === "oldValue") return `oldValue-${segment.value}`;
  return segment.type;
};

const getSegmentContent = ({
  activityLabels,
  field,
  fieldMeta,
  segment,
}: Omit<ActivityUpdateSegmentsProps, "currentValue" | "segments"> & {
  segment: Exclude<ActivityCopySegment, { type: "text" }>;
}) => {
  if (segment.type === "fieldLabel") {
    return isAssociationActivityField(field) ? (
      <ActivityAssociationBadge field={field} label={fieldMeta.label} />
    ) : (
      fieldMeta.label
    );
  }

  if (segment.type === "oldValue" && isAssociationActivityField(field)) {
    return <ActivityAssociationBadge field={field} label={segment.value} />;
  }

  if (segment.type === "currentValue" && field === "labels") {
    return <ActivityLabelValue labels={activityLabels} />;
  }

  return fieldMeta.render(segment.value);
};

export const ActivityUpdateSegments = ({
  activityLabels,
  currentValue,
  field,
  fieldMeta,
  segments,
}: ActivityUpdateSegmentsProps) => (
  <>
    {segments.map((segment) => {
      if (segment.type === "text") {
        return (
          <Text
            as="span"
            className="shrink-0 text-sm md:text-[0.95rem]"
            color="muted"
            key={getSegmentKey(segment)}
          >
            {segment.text}
          </Text>
        );
      }

      const isTruncatableValue =
        segment.type === "currentValue" &&
        (field === "title" || isAssociationActivityField(field));

      return (
        <Text
          as="span"
          className={cn(
            "inline-block text-sm text-black md:text-[0.95rem] dark:text-white",
            isTruncatableValue ? "min-w-0 truncate" : "shrink-0",
          )}
          fontWeight="medium"
          key={getSegmentKey(segment)}
          title={isTruncatableValue ? currentValue : undefined}
        >
          {getSegmentContent({
            activityLabels,
            field,
            fieldMeta,
            segment,
          })}
        </Text>
      );
    })}
  </>
);
