import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { aiChatKeys } from "../constants";
import { getTotalMessagesForTheMonth } from "../queries/get-total-messages";

export const useTotalMessages = () => {
  const { data: session } = useSession();

  return useQuery({
    queryKey: aiChatKeys.totalMessages(),
    queryFn: () => getTotalMessagesForTheMonth(session!),
    enabled: Boolean(session),
  });
};
