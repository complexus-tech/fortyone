"use client";
import { useEffect } from "react";
import { useAnalytics, useLocalStorage } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useSession } from "@/lib/auth/client";

export const IdentifyUser = () => {
  const { data: session } = useSession();
  const { analytics } = useAnalytics();

  const [lastIdentifyTime, setLastIdentifyTime] = useLocalStorage(
    "analytics:last-identify-time",
    0,
  );

  useEffect(() => {
    if (session) {
      const now = Date.now();
      const sixHours = DURATION_FROM_MILLISECONDS.HOUR * 6;

      // Only identify if 6 hours have passed since last identification
      if (now - lastIdentifyTime > sixHours) {
        const { email, name } = session.user;
        analytics.identify(email, {
          email,
          name,
        });
        setLastIdentifyTime(now);
      }
    }
  }, [session, analytics, lastIdentifyTime, setLastIdentifyTime]);

  return null;
};
