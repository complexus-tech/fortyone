import { useQuery } from "@tanstack/react-query";
import { linkKeys } from "@/constants/keys";
import { getLinkMetadata } from "../queries/links/get-metadata";

export const useLinkMetadata = (url: string) => {
  return useQuery({
    queryKey: linkKeys.metadata(url),
    queryFn: () => getLinkMetadata(url),
  });
};
