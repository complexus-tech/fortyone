import type { AppNotification } from "../types";

export type TemplateSegment =
  | {
      kind: "text";
      value: string;
    }
  | {
      emphasized: boolean;
      key: string;
      kind: "variable";
      value: string;
    };

type TemplateResult = {
  segments: TemplateSegment[];
  text: string;
};

export const renderTemplate = (
  message: AppNotification["message"],
): TemplateResult => {
  const { template, variables } = message;
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
    const variable = Object.prototype.hasOwnProperty.call(variables, key)
      ? variables[key]
      : undefined;

    segments.push(
      variable?.value
        ? {
            emphasized: variable.type !== "text",
            key,
            kind: "variable",
            value: variable.value,
          }
        : { kind: "text", value: token },
    );
    cursor = match.index + token.length;
    match = variablePattern.exec(template);
  }

  if (cursor < template.length) {
    segments.push({ kind: "text", value: template.slice(cursor) });
  }

  return {
    segments,
    text: segments.map(({ value }) => value).join(""),
  };
};
