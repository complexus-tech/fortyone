import type { ApiResponse } from "@/types/api-response";
import { auth } from "@/auth";
import { put } from "@/lib/http";
import { getApiError } from "@/utils/api-error";
import type {
  OnboardingTourProgress,
  UpdateOnboardingTourProgress,
} from "./types";
import { normalizeOnboardingTourProgress } from "./normalize-progress";

export const updateOnboardingTourProgressAction = async (
  payload: UpdateOnboardingTourProgress,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const response = await put<
      UpdateOnboardingTourProgress,
      ApiResponse<OnboardingTourProgress>
    >("onboarding/walkthrough", payload, { session: session!, workspaceSlug });

    return normalizeOnboardingTourProgress(response.data, {
      tourKey: payload.tourKey,
      tourVersion: payload.tourVersion,
    });
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(
      apiError.error?.message || "Failed to save onboarding progress",
    );
  }
};
