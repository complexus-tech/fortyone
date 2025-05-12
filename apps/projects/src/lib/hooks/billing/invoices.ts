import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getInvoices } from "@/lib/queries/billing/invoices";

export const useInvoices = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["invoices"],
    queryFn: () => getInvoices(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
