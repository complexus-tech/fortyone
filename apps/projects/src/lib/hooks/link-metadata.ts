import { useQuery } from "@tanstack/react-query";
import { linkKeys } from "@/constants/keys";
import { getLinkMetadata } from "../queries/links/get-metadata";

export const useLinkMetadata = (
  url: string,
  { enabled = true }: { enabled?: boolean } = {},
) => {
  return useQuery({
    enabled,
    queryKey: linkKeys.metadata(url),
    queryFn: () => getLinkMetadata(url),
  });
};
