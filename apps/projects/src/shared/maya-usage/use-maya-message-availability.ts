import { useSession } from "@/lib/auth/client";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useTotalMessages } from "./use-total-messages";
import { shouldShowMayaMessageLimit } from "./message-limit";

export const useMayaMessageAvailability = () => {
  const { data: totalMessages = 0, isError, isPending } = useTotalMessages();
  const { data: session } = useSession();
  const { getLimit } = useSubscriptionFeatures();

  return {
    isLimited:
      !isPending &&
      !isError &&
      shouldShowMayaMessageLimit({
        isInternalUser: session?.user.isInternal === true,
        limit: getLimit("maxAiMessages"),
        totalMessages,
      }),
    isPending,
    isUnavailable: isError,
  };
};
