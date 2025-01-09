"use server";
import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";
import type { ApiResponse } from "@/types";
import type { DocumentModel } from "../types";

export const getDocuments = async () => {
  const documents = await get<ApiResponse<DocumentModel[]>>("documents", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [TAGS.documents],
    },
  });
  return documents.data;
};
