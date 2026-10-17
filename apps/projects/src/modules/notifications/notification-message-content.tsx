import { Fragment } from "react";
import type { TemplateSegment } from "./utils/render-template";

const EMPHASIZED_VARIABLE_CLASS_NAME =
  "font-semibold antialiased text-foreground/90";

type NotificationMessageContentProps = {
  segments: TemplateSegment[];
  storyTerm: string;
};

export const NotificationMessageContent = ({
  segments,
  storyTerm,
}: NotificationMessageContentProps) => {
  return segments.map((segment, index) => {
    // Only localize producer-owned copy; names, titles, and comments are data.
    const value =
      segment.kind === "text"
        ? segment.value.replace(/\bstory\b/g, () => storyTerm)
        : segment.value;
    const key = `${segment.kind}-${"key" in segment ? segment.key : "text"}-${index}`;

    if (segment.kind === "variable" && segment.emphasized) {
      return (
        <span className={EMPHASIZED_VARIABLE_CLASS_NAME} key={key}>
          {value}
        </span>
      );
    }

    return <Fragment key={key}>{value}</Fragment>;
  });
};
