"use server";
import { linkTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import ky from "ky";

export type LinkMetadata = {
  title?: string;
  description?: string;
  image?: string;
};

export const getLinkMetadata = async (url: string) => {
  const metadata = await ky
    .get(`https://api.dub.co/metatags?url=${url}`, {
      next: {
        revalidate: DURATION_FROM_SECONDS.DAY * 3,
        tags: [linkTags.metadata(url)],
      },
    })
    .json<LinkMetadata>();
  return metadata;
};
