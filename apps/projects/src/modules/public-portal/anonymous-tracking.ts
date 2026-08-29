import type { PublicPortal, PublicRequest } from "./types";
import { getRequestPath } from "./utils";

export const getAnonymousFeedbackTrackingUrl = (
  portal: PublicPortal,
  request: PublicRequest,
) =>
  new URL(getRequestPath(portal, request), window.location.origin).toString();
