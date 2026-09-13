import * as Sentry from "@sentry/react-native";
import { scrubErrorEvent } from "./observability-privacy";

let initialized = false;
const capturedErrors = new WeakSet<Error>();

export const initializeObservability = () => {
  const dsn = process.env.EXPO_PUBLIC_SENTRY_DSN?.trim();
  if (initialized || __DEV__ || !dsn) return;

  Sentry.init({
    dsn,
    sendDefaultPii: false,
    enableLogs: false,
    enableAutoSessionTracking: false,
    enableAutoPerformanceTracing: false,
    enableNativeFramesTracking: false,
    enableAppHangTracking: false,
    enableCaptureFailedRequests: false,
    attachScreenshot: false,
    attachViewHierarchy: false,
    maxBreadcrumbs: 0,
    beforeBreadcrumb: () => null,
    beforeSend: (event, hint) => {
      // Attachments bypass event fields; do not send screen captures or files.
      hint.attachments = [];
      return scrubErrorEvent(event);
    },
    // Traces, profiles, and replay are intentionally unconfigured. Setting their
    // sample rates to zero still installs some SDK instrumentation.
  });
  initialized = true;
};

export const captureAppError = (error: Error) => {
  if (!initialized || capturedErrors.has(error)) return;
  capturedErrors.add(error);
  Sentry.captureException(error);
};
