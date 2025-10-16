import { post } from "@/lib/http/fetch";
import type { ApiResponse, Label } from "@/types";

export type NewLabel = {
  name: string;
  color: string;
  teamId?: string;
};

export const createLabel = async (newLabel: NewLabel) => {
  const response = await post<NewLabel, ApiResponse<Label>>("labels", newLabel);
  return response;
};
