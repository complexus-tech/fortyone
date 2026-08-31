import { auth } from "@/auth";
import { put } from "@/lib/http";
import { normalizeOnboardingTourProgress } from "@/lib/onboarding-tour-progress";
import type {
  ApiResponse,
  OnboardingTourProgress,
  UpdateOnboardingTourProgress,
} from "@/types";
import { getApiError } from "@/utils";

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
