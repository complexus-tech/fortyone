"use client";

import { useEffect } from "react";
import { useSession } from "next-auth/react";
import { WalkthroughProvider, useWalkthrough } from "./walkthrough-provider";
import { WalkthroughOverlay } from "./walkthrough-overlay";
import { useWalkthroughSteps } from "./use-walkthrough-steps";

const WalkthroughManager = () => {
  const { data: session } = useSession();
  const { setSteps, startWalkthrough, state } = useWalkthrough();
  const steps = useWalkthroughSteps();

  useEffect(() => {
    setSteps(steps);
  }, [steps, setSteps]);

  // Auto-start walkthrough for new users
  useEffect(() => {
    if (session && !state.hasSeenWalkthrough && steps.length > 0) {
      // Small delay to ensure DOM is ready
      const timeout = setTimeout(() => {
        startWalkthrough();
      }, 1000);

      return () => {
        clearTimeout(timeout);
      };
    }
  }, [session, state.hasSeenWalkthrough, steps.length, startWalkthrough]);

  return <WalkthroughOverlay />;
};

export const WalkthroughIntegration = () => {
  return (
    <WalkthroughProvider autoStart={false} version="1.0.0">
      <WalkthroughManager />
    </WalkthroughProvider>
  );
};
