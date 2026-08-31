import { auth } from "@/auth";
import { put } from "@/lib/http";
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

    return response.data!;
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(
      apiError.error?.message || "Failed to save onboarding progress",
    );
  }
};
