import type {
  FeedbackWidgetMode,
  FeedbackWidgetTab,
  FeedbackWidgetTheme,
} from "./protocol";

export type FeedbackWidgetInstallOptions = {
  defaultTab?: FeedbackWidgetTab;
  mode?: FeedbackWidgetMode;
  portalSlug: string;
  position?: "bottom-left" | "bottom-right";
  scriptOrigin?: string;
  theme?: FeedbackWidgetTheme;
  trigger?: string;
};

const DEFAULT_WIDGET_ORIGIN = "https://cloud.fortyone.app";

const escapeAttribute = (value: string) =>
  value
    .replaceAll("&", "&amp;")
    .replaceAll('"', "&quot;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");

export const buildFeedbackWidgetSnippet = ({
  defaultTab = "feedback",
  mode = "bubble",
  portalSlug,
  position = "bottom-right",
  scriptOrigin = DEFAULT_WIDGET_ORIGIN,
  theme = "auto",
  trigger,
}: FeedbackWidgetInstallOptions) => {
  const origin = new URL(scriptOrigin).origin;
  const resolvedDefaultTab = defaultTab === "updates" ? "feedback" : defaultTab;
  const attributes = [
    ["src", `${origin}/api/feedback-widget/v1.js`],
    ["data-portal", portalSlug],
    ["data-mode", mode],
    ["data-default-tab", resolvedDefaultTab],
    ["data-theme", theme],
    ["data-position", position],
    ...(mode === "custom" && trigger ? [["data-trigger", trigger]] : []),
  ];
  const serializedAttributes = attributes
    .map(([name, value]) => `${name}="${escapeAttribute(value)}"`)
    .join("\n  ");

  return `<script
  ${serializedAttributes}
  async
></script>`;
};

export const buildInlineFeedbackWidgetMarkup = (
  options: Omit<FeedbackWidgetInstallOptions, "mode">,
) => {
  const targetId = "fortyone-feedback";
  return `<div id="${targetId}"></div>\n${buildFeedbackWidgetSnippet({
    ...options,
    mode: "inline",
    trigger: `#${targetId}`,
  }).replace(
    'data-mode="inline"',
    `data-mode="inline"\n  data-target="#${targetId}"`,
  )}`;
};
