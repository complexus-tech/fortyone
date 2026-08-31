import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, OnboardingTourProgress } from "@/types";

type OnboardingTourProgressQuery = {
  tourKey: string;
  tourVersion: string;
};

export const getOnboardingTourProgress = async (
  ctx: WorkspaceCtx,
  { tourKey, tourVersion }: OnboardingTourProgressQuery,
) => {
  const query = new URLSearchParams({ tourKey, tourVersion });
  const progress = await get<ApiResponse<OnboardingTourProgress>>(
    `onboarding/walkthrough?${query.toString()}`,
    ctx,
  );

  return progress.data!;
};
