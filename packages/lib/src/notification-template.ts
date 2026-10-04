import { htmlToPlainText } from "./html-to-plain-text";

type NotificationVariable = { value: string; type?: string };
type NotificationMessage = {
  template: string;
  variables: Record<string, NotificationVariable> | null;
};

export type TemplateSegment =
  | { kind: "text"; value: string }
  | { kind: "variable"; key: string; value: string; emphasized: boolean };

export const renderNotificationTemplate = ({
  template,
  variables,
}: NotificationMessage): { segments: TemplateSegment[]; text: string } => {
  const segments: TemplateSegment[] = [];
  const variablePattern = /\{\w+\}/g;
  let cursor = 0;
  let match = variablePattern.exec(template);

  while (match) {
    if (match.index > cursor) {
      segments.push({
        kind: "text",
        value: template.slice(cursor, match.index),
      });
    }
    const token = match[0];
    const key = token.slice(1, -1);
    const variable =
      variables && Object.prototype.hasOwnProperty.call(variables, key)
        ? variables[key]
        : undefined;
    if (variable && typeof variable.value === "string") {
      const isContent = variable.type === "text" || key === "content";
      segments.push({
        kind: "variable",
        key,
        emphasized: !isContent && variable.type !== "plain_text",
        value:
          isContent && variable.type !== "plain_text"
            ? htmlToPlainText(variable.value)
            : variable.value,
      });
    } else {
      segments.push({ kind: "text", value: token });
    }
    cursor = match.index + token.length;
    match = variablePattern.exec(template);
  }
  if (cursor < template.length) {
    segments.push({ kind: "text", value: template.slice(cursor) });
  }
  return { segments, text: segments.map(({ value }) => value).join("") };
};
