import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { NewObjective, Objective } from "../types";

export const createObjective = async (params: NewObjective) => {
  try {
    const res = await post<NewObjective, ApiResponse<Objective>>(
      "objectives",
      params,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
