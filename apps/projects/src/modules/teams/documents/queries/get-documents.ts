"use server";
import { get } from "@/lib/http";
import { DocumentModel } from "../types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";

export const getDocuments = async () => {
  const documents = await get<DocumentModel[]>("documents", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [TAGS.documents],
    },
  });
  return documents;
};
