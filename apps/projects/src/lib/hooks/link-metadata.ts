import {
  useIsMutating,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { linkKeys } from "@/constants/keys";
import type { LinkMetadata } from "../queries/links/metadata";
import { getLinkMetadata } from "../queries/links/get-metadata";
import {
  hasLinkMetadata,
  METADATA_FAILURE_TTL_MS,
  METADATA_SUCCESS_TTL_MS,
  normalizeMetadataUrl,
} from "../queries/links/metadata";

export const useLinkMetadata = (
  url: string,
  { enabled = true }: { enabled?: boolean } = {},
) => {
  const normalizedUrl = normalizeMetadataUrl(url) ?? "";
  const queryClient = useQueryClient();
  const refreshKey = ["refresh-link-preview", normalizedUrl];
  const refreshMutation = useMutation({
    mutationKey: refreshKey,
    mutationFn: (target: string) =>
      getLinkMetadata(target, undefined, { refresh: true }),
    onSuccess: (metadata, target) => {
      queryClient.setQueryData(linkKeys.metadata(target), metadata);
    },
  });
  const isRefreshing = useIsMutating({ mutationKey: refreshKey }) > 0;
  const query = useQuery({
    enabled: enabled && Boolean(normalizedUrl),
    queryKey: linkKeys.metadata(normalizedUrl),
    queryFn: async ({ signal }) => {
      const data = await getLinkMetadata(normalizedUrl, signal);
      // Failed optional enrichment must not erase a previously useful preview.
      const previous = queryClient.getQueryData<LinkMetadata>(
        linkKeys.metadata(normalizedUrl),
      );
      if (!hasLinkMetadata(data) && hasLinkMetadata(previous))
        return { ...previous!, stale: false };
      return data;
    },
    refetchInterval: (query) => (query.state.data?.stale ? 2000 : false),
    staleTime: (query) => {
      if (query.state.data?.stale) return 0;
      return hasLinkMetadata(query.state.data)
        ? METADATA_SUCCESS_TTL_MS
        : METADATA_FAILURE_TTL_MS;
    },
    gcTime: 24 * 60 * 60 * 1000,
    retry: false,
  });
  return {
    ...query,
    isRefreshing,
    refreshError: refreshMutation.error,
    refreshPreview: async () => {
      if (!normalizedUrl || isRefreshing) return;
      await queryClient.cancelQueries({
        queryKey: linkKeys.metadata(normalizedUrl),
      });
      await refreshMutation.mutateAsync(normalizedUrl);
    },
  };
};
