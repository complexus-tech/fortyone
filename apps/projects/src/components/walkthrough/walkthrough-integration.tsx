"use client";

import { useEffect } from "react";
import { WalkthroughProvider, useWalkthrough } from "./walkthrough-provider";
import { WalkthroughOverlay } from "./walkthrough-overlay";
import { useWalkthroughSteps } from "./use-walkthrough-steps";

const WalkthroughManager = () => {
  const { setSteps } = useWalkthrough();
  const steps = useWalkthroughSteps();

  useEffect(() => {
    setSteps(steps);
  }, [steps, setSteps]);

  return <WalkthroughOverlay />;
};

export const WalkthroughIntegration = () => {
  return (
    <WalkthroughProvider autoStart version="1.0.0">
      <WalkthroughManager />
    </WalkthroughProvider>
  );
};
