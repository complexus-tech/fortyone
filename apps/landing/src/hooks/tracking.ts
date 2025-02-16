import { usePostHog } from "posthog-js/react";
import type { Analytics } from "@/types/analytics";

export const useAnalytics = () => {
  const posthog = usePostHog();

  const analytics: Analytics = {
    track: (eventName, properties) => {
      posthog.capture(eventName, properties);
    },
    startSessionRecording: () => {
      posthog.startSessionRecording();
    },
    stopSessionRecording: () => {
      posthog.stopSessionRecording();
    },
    identify: (userId, properties, propertiesToSetOnce) => {
      posthog.identify(userId, properties, propertiesToSetOnce);
    },
    logout: () => {
      posthog.reset();
    },
  };

  return { analytics };
};
