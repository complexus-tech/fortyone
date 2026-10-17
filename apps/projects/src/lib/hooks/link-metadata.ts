import {
  hashKey,
  useIsMutating,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useMemo } from "react";
import { linkKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { LinkMetadata } from "../queries/links/metadata";
import { getLinkMetadata } from "../queries/links/get-metadata";
import {
  hasLinkMetadata,
  METADATA_FAILURE_TTL_MS,
  normalizeMetadataUrl,
} from "../queries/links/metadata";
import {
  PREVIEW_CACHE_TTL_MS,
  readPreviewCache,
  usePreviewCacheReady,
  writePreviewCache,
} from "../queries/links/preview-cache";

const isLinkMetadata = (value: unknown): value is LinkMetadata => {
  if (typeof value !== "object" || value === null) return false;
  const metadata = value as Record<string, unknown>;
  const fields = ["title", "description", "image", "previewImage"];
  return (
    fields.every(
      (field) =>
        metadata[field] === undefined || typeof metadata[field] === "string",
    ) && fields.some((field) => Boolean(metadata[field]))
  );
};

export const useLinkMetadata = (
  url: string,
  { enabled = true }: { enabled?: boolean } = {},
) => {
  const normalizedUrl = normalizeMetadataUrl(url) ?? "";
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const identity = session?.user.id;
  const cacheReady = usePreviewCacheReady();
  const queryClient = useQueryClient();
  const queryKey = [
    ...linkKeys.metadata(normalizedUrl),
    identity,
    workspaceSlug,
  ];
  const cacheKey = hashKey(queryKey);
  const persisted = useMemo(
    () =>
      identity && workspaceSlug
        ? readPreviewCache(cacheKey, isLinkMetadata)
        : undefined,
    [cacheKey, identity, workspaceSlug],
  );
  const persist = (metadata: LinkMetadata) => {
    if (
      identity &&
      workspaceSlug &&
      hasLinkMetadata(metadata) &&
      !metadata.stale
    )
      writePreviewCache(cacheKey, metadata);
  };
  const refreshKey = [
    "refresh-link-preview",
    normalizedUrl,
    identity,
    workspaceSlug,
  ];
  const refreshMutation = useMutation({
    mutationKey: refreshKey,
    mutationFn: async (request: {
      url: string;
      queryKey: typeof queryKey;
      cacheKey: string;
      persist: boolean;
    }) => {
      const metadata = await getLinkMetadata(request.url, undefined, {
        refresh: true,
      });
      if (!hasLinkMetadata(metadata))
        throw new Error("This preview could not be refreshed.");
      return metadata;
    },
    onSuccess: (metadata, request) => {
      queryClient.setQueryData(request.queryKey, metadata);
      if (request.persist && !metadata.stale)
        writePreviewCache(request.cacheKey, metadata);
    },
  });
  const isRefreshing = useIsMutating({ mutationKey: refreshKey }) > 0;
  const query = useQuery({
    enabled: enabled && Boolean(normalizedUrl) && Boolean(identity),
    queryKey,
    initialData: persisted?.data,
    initialDataUpdatedAt: persisted?.savedAt,
    queryFn: async ({ signal }) => {
      const data = await getLinkMetadata(normalizedUrl, signal);
      // Failed optional enrichment must not erase a previously useful preview.
      const previous = queryClient.getQueryData<LinkMetadata>(queryKey);
      if (!hasLinkMetadata(data) && hasLinkMetadata(previous))
        return { ...previous!, stale: false };
      if (!signal.aborted) persist(data);
      return data;
    },
    refetchInterval: (query) => (query.state.data?.stale ? 2000 : false),
    staleTime: (query) => {
      if (query.state.data?.stale) return 0;
      return hasLinkMetadata(query.state.data)
        ? PREVIEW_CACHE_TTL_MS
        : METADATA_FAILURE_TTL_MS;
    },
    gcTime: PREVIEW_CACHE_TTL_MS,
    retry: false,
  });
  return {
    ...query,
    data: cacheReady ? query.data : undefined,
    isRefreshing,
    refreshError: refreshMutation.error,
    refreshPreview: async () => {
      if (!normalizedUrl || !identity || isRefreshing) return;
      await queryClient.cancelQueries({ queryKey });
      await refreshMutation.mutateAsync({
        url: normalizedUrl,
        queryKey,
        cacheKey,
        persist: Boolean(identity && workspaceSlug),
      });
    },
  };
};
