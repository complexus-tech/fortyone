"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import {
  WalkthroughProvider,
  useWalkthrough,
  type WalkthroughStep,
} from "./walkthrough-provider";
import { WalkthroughOverlay } from "./walkthrough-overlay";
import { useWalkthroughSteps } from "./use-walkthrough-steps";
import { getModuleWalkthroughTour } from "./module-walkthrough-tours";

const WalkthroughManager = ({ steps }: { steps: WalkthroughStep[] }) => {
  const { setSteps } = useWalkthrough();

  useEffect(() => {
    setSteps(steps);
  }, [steps, setSteps]);

  return <WalkthroughOverlay />;
};

export const WalkthroughIntegration = () => {
  const pathname = usePathname();
  const { workspaceSlug } = useWorkspacePath();
  const { state } = useWalkthrough();
  const gettingStartedSteps = useWalkthroughSteps();
  const moduleTour = getModuleWalkthroughTour(pathname, workspaceSlug);
  const activeModuleTour =
    state.hasSeenWalkthrough && !state.isActive ? moduleTour : null;

  return (
    <>
      <WalkthroughManager steps={gettingStartedSteps} />
      {activeModuleTour ? (
        <WalkthroughProvider
          autoStart
          dismissOnClose
          key={`${activeModuleTour.tourKey}:${activeModuleTour.version}`}
          tourKey={activeModuleTour.tourKey}
          version={activeModuleTour.version}
        >
          <WalkthroughManager steps={activeModuleTour.steps} />
        </WalkthroughProvider>
      ) : null}
    </>
  );
};
