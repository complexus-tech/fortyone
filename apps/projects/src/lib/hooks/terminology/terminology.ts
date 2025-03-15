import { useQuery } from "@tanstack/react-query";
import { terminologyKeys } from "@/constants/keys";
import { getTerminology } from "../../queries/terminology/get-terminology";

export const useTerminology = () => {
  return useQuery({
    queryKey: terminologyKeys.all,
    queryFn: getTerminology,
  });
};
