import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import { get } from "@/lib/http";
import type { OnboardingTourProgress } from "./types";
import { normalizeOnboardingTourProgress } from "./normalize-progress";

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

  return normalizeOnboardingTourProgress(progress.data, {
    tourKey,
    tourVersion,
  });
};
