"use server";

import { put } from "@/lib/http";
import { ApiResponse, Link } from "@/types";

type Payload = {
  url: string;
  title?: string;
};

export const updateLinkAction = async (linkId: string, payload: Payload) => {
  const _ = await put<Payload, ApiResponse<Link>>(`links/${linkId}`, payload);
  return linkId;
};
