import React, { Fragment } from "react";
import { htmlToPlainText } from "lib/src/html-to-plain-text";
import { renderNotificationTemplate } from "lib/src/notification-template";
import { Text } from "@/components/ui";
import type { AppNotification } from "../types";

export const htmlToText = htmlToPlainText;
export const renderTemplate = renderNotificationTemplate;

export const renderTemplateJSX = (
  message: AppNotification["message"],
  storyTerm: string,
): React.ReactElement => {
  const { segments } = renderTemplate({
    ...message,
    template: message.template.replace("story", storyTerm),
  });
  return (
    <>
      {segments.map((segment, index) => {
        const key = `${segment.kind}-${"key" in segment ? segment.key : "text"}-${index}`;
        return segment.kind === "variable" && segment.emphasized ? (
          <Text key={key} color="muted" fontSize="sm" fontWeight="medium">
            {segment.value}
          </Text>
        ) : (
          <Fragment key={key}>{segment.value}</Fragment>
        );
      })}
    </>
  );
};
