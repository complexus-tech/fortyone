import { getTrustedWidgetOrigin } from "./protocol";

export const getFeedbackWidgetFrameAncestors = (
  parentOrigin: string | null,
) => {
  const trustedOrigin = getTrustedWidgetOrigin(parentOrigin);
  return trustedOrigin &&
    trustedOrigin === parentOrigin &&
    !trustedOrigin.includes("*")
    ? `frame-ancestors ${trustedOrigin}`
    : "frame-ancestors 'none'";
};

export const isValidFeedbackWidgetParent = (parentOrigin: string | null) =>
  getFeedbackWidgetFrameAncestors(parentOrigin) !== "frame-ancestors 'none'";
