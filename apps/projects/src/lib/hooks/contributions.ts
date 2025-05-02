import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getContributions } from "../queries/analytics/get-contributions";

export const useContributions = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["contributions"],
    queryFn: () => getContributions(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
