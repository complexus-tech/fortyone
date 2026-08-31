"use client";

import { useEffect } from "react";
import { useWalkthrough } from "./walkthrough-provider";
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

export const WalkthroughIntegration = () => <WalkthroughManager />;
