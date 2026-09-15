import { useQuery } from "@tanstack/react-query";
import { memberKeys } from "@/constants/keys";
import { getMayaAssignee } from "../queries/get-maya-assignee";

export function useMayaAssignee() {
  return useQuery({
    queryKey: memberKeys.maya(),
    queryFn: ({ signal }) => getMayaAssignee(signal),
    staleTime: 10 * 60 * 1000,
  });
}
