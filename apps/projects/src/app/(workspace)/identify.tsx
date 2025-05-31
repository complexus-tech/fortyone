"use client";
import { useSession } from "next-auth/react";
import { useEffect } from "react";
import { useAnalytics, useLocalStorage } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

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
      const twentyFourHours = DURATION_FROM_MILLISECONDS.HOUR * 24;

      // Only identify if 24 hours have passed since last identification
      if (now - lastIdentifyTime > twentyFourHours) {
        analytics.identify(session.user!.email!, {
          email: session.user!.email!,
          name: session.user!.name!,
        });
        setLastIdentifyTime(now);
      }
    }
  }, [session, analytics, lastIdentifyTime, setLastIdentifyTime]);

  return null;
};
