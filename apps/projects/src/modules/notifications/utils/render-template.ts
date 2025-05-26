import type { AppNotification } from "../types";

type TemplateResult = {
  text: string;
  html: string;
};

export const renderTemplate = (
  message: AppNotification["message"],
): TemplateResult => {
  const { template, variables } = message;

  // Helper function to replace variables in template
  const replaceVariables = (template: string, wrapInSpan = false): string => {
    return template.replace(/\{\w+\}/g, (match) => {
      const key = match.slice(1, -1); // Remove { and }
      const variable = variables[key as keyof typeof variables];

      if (variable.value) {
        if (wrapInSpan) {
          return `<span class="font-semibold antialiased text-black dark:text-gray-200">${variable.value}</span>`;
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
