import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { KeyResult, NewObjective, Objective } from "../types";

export type CreateObjectiveResponse = {
  objective: Objective;
  keyResults?: KeyResult[];
};

export const createObjective = async (
  params: NewObjective,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<NewObjective, ApiResponse<CreateObjectiveResponse>>(
      "objectives",
      params,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
