import { useQuery } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { userKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { useSession } from "@/lib/auth/client";
import { getOnboardingTourProgress } from "@/lib/queries/users/onboarding-tour-progress";

type OnboardingTourProgressOptions = {
  tourKey: string;
  tourVersion: string;
};

export const useOnboardingTourProgress = ({
  tourKey,
  tourVersion,
}: OnboardingTourProgressOptions) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: userKeys.onboardingTourProgress(
      workspaceSlug,
      tourKey,
      tourVersion,
    ),
    queryFn: () =>
      getOnboardingTourProgress(
        { session: session!, workspaceSlug },
        { tourKey, tourVersion },
      ),
    enabled: Boolean(session && workspaceSlug),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};
