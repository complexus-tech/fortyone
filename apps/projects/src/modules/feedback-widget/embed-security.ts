import { getTrustedWidgetOrigin } from "./protocol";

export type PublicFeedbackWidgetConfig = {
  allowedOrigins: string[];
  enabled: boolean;
  widgetKeyId: string;
};

export const getFeedbackWidgetFrameAncestors = (origins: string[]) => {
  const trusted = origins.filter(
    (origin) => getTrustedWidgetOrigin(origin) === origin,
  );
  return trusted.length > 0
    ? `frame-ancestors ${trusted.join(" ")}`
    : "frame-ancestors 'none'";
};

export const isAllowedFeedbackWidgetParent = (
  config: PublicFeedbackWidgetConfig,
  parentOrigin: string | null,
) => {
  if (!config.enabled || !parentOrigin) return false;
  return (
    getTrustedWidgetOrigin(parentOrigin) === parentOrigin &&
    config.allowedOrigins.includes(parentOrigin)
  );
};
