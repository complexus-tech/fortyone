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
  defaultTab = "home",
  mode = "bubble",
  portalSlug,
  position = "bottom-right",
  scriptOrigin = DEFAULT_WIDGET_ORIGIN,
  theme = "auto",
  trigger,
}: FeedbackWidgetInstallOptions) => {
  const origin = new URL(scriptOrigin).origin;
  const attributes = [
    ["src", `${origin}/api/feedback-widget/v1.js`],
    ["data-portal", portalSlug],
    ["data-mode", mode],
    ["data-default-tab", defaultTab],
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

export const buildFeedbackWidgetIdentityServerExample = ({
  origin,
  signingSecretVersion,
  widgetKeyId,
}: {
  origin: string;
  signingSecretVersion: number;
  widgetKeyId: string;
}) => `import { createHmac, randomUUID } from "node:crypto";

export function createFortyOneFeedbackIdentity(user) {
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    version: ${signingSecretVersion},
    keyId: ${JSON.stringify(widgetKeyId)},
    externalId: user.id,
    email: user.email,
    displayName: user.name,
    avatarUrl: user.avatarUrl,
    iat: now,
    exp: now + 4 * 60,
    nonce: randomUUID(),
    origin: ${JSON.stringify(origin)},
  };
  const encoded = Buffer.from(JSON.stringify(payload)).toString("base64url");
  const signature = createHmac(
    "sha256",
    process.env.FORTYONE_FEEDBACK_WIDGET_SECRET,
  ).update(encoded).digest("base64url");
  return \`\${encoded}.\${signature}\`;
}`;
