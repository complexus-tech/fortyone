import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { labelKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getLabels } from "../queries/labels/get-labels";

export const useLabels = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: labelKeys.lists(),
    queryFn: () => getLabels(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
