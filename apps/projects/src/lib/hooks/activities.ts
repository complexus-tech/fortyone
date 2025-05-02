import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getActivities } from "../queries/activities/get-activities";

export const useActivities = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["activities"],
    queryFn: () => getActivities(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
