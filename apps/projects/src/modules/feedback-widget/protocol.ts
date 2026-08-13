export const FEEDBACK_WIDGET_CHANNEL = "fortyone:feedback-widget";
export const FEEDBACK_WIDGET_VERSION = 1;

export type FeedbackWidgetTab = "feedback" | "roadmap" | "updates";
export type FeedbackWidgetTheme = "auto" | "light" | "dark";
export type FeedbackWidgetMode = "bubble" | "custom" | "inline";

export type FeedbackWidgetFrameEvent =
  | "close"
  | "escape"
  | "identity-cleared"
  | "identity-error"
  | "identity-ready"
  | "open-external"
  | "ready"
  | "resize";

export type FeedbackWidgetHostEvent =
  | "host-close"
  | "host-identify"
  | "host-identity-clear"
  | "host-open";

export type FeedbackWidgetMessage = {
  channel: typeof FEEDBACK_WIDGET_CHANNEL;
  event: FeedbackWidgetFrameEvent | FeedbackWidgetHostEvent;
  instanceId: string;
  payload?: Record<string, unknown>;
  version: typeof FEEDBACK_WIDGET_VERSION;
};

export const getTrustedWidgetOrigin = (value?: string | null) => {
  if (!value) return null;

  try {
    const url = new URL(value);
    const isLoopback =
      url.hostname === "localhost" ||
      url.hostname === "127.0.0.1" ||
      url.hostname === "[::1]";
    if (
      url.protocol !== "https:" &&
      !(url.protocol === "http:" && isLoopback)
    ) {
      return null;
    }
    if (url.username || url.password || url.origin !== value) return null;
    return url.origin;
  } catch {
    return null;
  }
};

export const isFeedbackWidgetMessage = (
  value: unknown,
  instanceId: string,
): value is FeedbackWidgetMessage => {
  if (!value || typeof value !== "object") return false;
  const message = value as Partial<FeedbackWidgetMessage>;

  return (
    message.channel === FEEDBACK_WIDGET_CHANNEL &&
    message.version === FEEDBACK_WIDGET_VERSION &&
    message.instanceId === instanceId &&
    typeof message.event === "string"
  );
};

export const postFeedbackWidgetMessage = (
  event: FeedbackWidgetFrameEvent,
  instanceId: string,
  parentOrigin: string,
  payload?: Record<string, unknown>,
) => {
  if (typeof window === "undefined" || window.parent === window) return;
  const trustedOrigin = getTrustedWidgetOrigin(parentOrigin);
  if (!trustedOrigin) return;

  const message: FeedbackWidgetMessage = {
    channel: FEEDBACK_WIDGET_CHANNEL,
    event,
    instanceId,
    payload,
    version: FEEDBACK_WIDGET_VERSION,
  };
  window.parent.postMessage(message, trustedOrigin);
};
