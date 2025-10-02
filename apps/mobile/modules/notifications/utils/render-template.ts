import type { AppNotification } from "../types";

export const renderTemplate = (message: AppNotification["message"]) => {
  let html = message.template;
  let text = message.template;

  // Replace variables in template
  Object.entries(message.variables).forEach(([key, variable]) => {
    const placeholder = `{{${key}}}`;
    html = html.replace(new RegExp(placeholder, "g"), variable.value);
    text = text.replace(new RegExp(placeholder, "g"), variable.value);
  });

  return { html, text };
};
