import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Invoice } from "@/types";

export const getInvoices = async (session: Session) => {
  try {
    const invoices = await get<ApiResponse<Invoice[]>>("invoices", session);
    return invoices.data!;
  } catch {
    return [];
  }
};
