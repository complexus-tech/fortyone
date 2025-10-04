import React from "react";
import { Text } from "@/components/ui";
import type { AppNotification } from "../types";

type TemplateResult = {
  text: string;
  html: string;
};

export function htmlToText(html: string): string {
  if (!html) return "";
  // Replace tags with nothing
  return html.replace(/<[^>]+>/g, "");
}

export const renderTemplate = (
  message: AppNotification["message"]
): TemplateResult => {
  const { template, variables } = message;

  // Helper function to replace variables in template
  const replaceVariables = (template: string, wrapInSpan = false): string => {
    return template.replace(/\{\w+\}/g, (match) => {
      const key = match.slice(1, -1); // Remove { and }
      const variable = variables[key as keyof typeof variables];

      if (variable.value) {
        if (wrapInSpan) {
          return `<span class="font-semibold antialiased text-black/80 dark:text-gray-200/95">${variable.value}</span>`;
        }
        return variable.value;
      }
      return match;
    });
  };

  const text = replaceVariables(template, false);
  const html = replaceVariables(template, true);

  return { text, html };
};

export const renderTemplateJSX = (
  message: AppNotification["message"],
  storyTerm: string
): React.ReactElement => {
  const { template, variables } = message;

  // Split template by variable placeholders and create JSX elements
  const parts = template.replace("story", storyTerm).split(/(\{\w+\})/g);

  const elements = parts
    .map((part, index) => {
      if (part.match(/^\{\w+\}$/)) {
        // This is a variable placeholder
        const key = part.slice(1, -1); // Remove { and }
        const variable = variables[key as keyof typeof variables];

        if (variable.value) {
          return (
            <Text key={index} fontSize="sm" fontWeight="semibold">
              {htmlToText(variable.value)}
            </Text>
          );
        }
        return part;
      }

      // This is regular text
      if (part.trim()) {
        return (
          <Text key={index} color="muted" fontSize="sm">
            {htmlToText(part)}
          </Text>
        );
      }

      return null;
    })
    .filter(Boolean);

  return <>{elements}</>;
};
