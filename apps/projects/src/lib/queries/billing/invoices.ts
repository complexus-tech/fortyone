import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Invoice } from "@/types";

export const getInvoices = async (ctx: WorkspaceCtx) => {
  try {
    const invoices = await get<ApiResponse<Invoice[]>>("invoices", ctx);
    return invoices.data!;
  } catch {
    return [];
  }
};
