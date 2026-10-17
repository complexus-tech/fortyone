"use client";

import type { QueryClient, QueryKey } from "@tanstack/react-query";
import {
  hashKey,
  skipToken,
  useQuery,
  useQueryClient,
  useMutationState,
} from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import type { WorkspaceCtx } from "@/lib/http";
import { get } from "@/lib/http";
import { useSession } from "@/lib/auth/client";
import { storyKeys } from "@/shared/story/cache-keys";
import type { DetailedStory } from "@/shared/story/types";
import { objectiveKeys } from "@/shared/objectives/keys";
import type { Objective } from "@/shared/objectives/types";
import { documentKeys } from "@/shared/documents/keys";
import type { WorkspaceDocument } from "@/modules/documents/types";
import type { ApiResponse } from "@/types/api-response";
import {
  PREVIEW_CACHE_TTL_MS,
  readPreviewCache,
  removePreviewCache,
  usePreviewCacheReady,
  writePreviewCache,
} from "@/lib/queries/links/preview-cache";
import type { NativeLinkTarget } from "./native-target";

export type NativeStorySummary = Pick<
  DetailedStory,
  | "id"
  | "title"
  | "teamId"
  | "teamCode"
  | "workspaceId"
  | "sequenceId"
  | "statusId"
  | "assigneeId"
  | "assignee"
  | "priority"
  | "updatedAt"
  | "archivedAt"
  | "deletedAt"
>;
export type NativeObjectiveSummary = Pick<
  Objective,
  | "id"
  | "name"
  | "sequenceId"
  | "shortSummary"
  | "teamId"
  | "workspaceId"
  | "leadUser"
  | "statusId"
  | "priority"
  | "updatedAt"
>;
export type NativeDocumentSummary = Pick<
  WorkspaceDocument,
  "id" | "title" | "contentText" | "workspaceId" | "revision" | "updatedAt"
>;
export type NativeLinkResource = (
  | { kind: "story"; value: NativeStorySummary }
  | { kind: "objective"; value: NativeObjectiveSummary }
  | { kind: "document"; value: NativeDocumentSummary }
) & { version?: string; refreshedAt?: string };

type CachedPreview = {
  resource: NativeLinkResource | null;
  requestedAt: number;
};

const isCachedPreview = (value: unknown): value is CachedPreview => {
  if (!isRecord(value) || typeof value.requestedAt !== "number") return false;
  const resource = value.resource;
  if (!isRecord(resource) || !isRecord(resource.value)) return false;
  if (typeof resource.value.id !== "string") return false;
  return resource.kind === "objective"
    ? typeof resource.value.name === "string"
    : (resource.kind === "story" || resource.kind === "document") &&
        typeof resource.value.title === "string";
};
// A query can have several card observers. The refresh signal must belong to
// the shared query, not the particular card whose menu was clicked.
const refreshRequests = new WeakMap<QueryClient, Set<string>>();
const requestRefresh = (client: QueryClient, queryHash: string) => {
  let requests = refreshRequests.get(client);
  if (!requests) {
    requests = new Set();
    refreshRequests.set(client, requests);
  }
  requests.add(queryHash);
};
const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null;

const applyWorkMutation = (
  resource: NativeLinkResource,
  variables: unknown,
): NativeLinkResource => {
  if (!isRecord(variables)) return resource;
  let patch: unknown;
  if (resource.kind === "story" && variables.storyId === resource.value.id)
    patch = variables.payload;
  else if (
    resource.kind === "objective" &&
    variables.objectiveId === resource.value.id
  )
    patch = variables.data;
  if (!isRecord(patch)) return resource;
  // Only presentation fields belong in a summary; never copy HTML, attachments
  // or mutation-only inputs into this cache.
  const fields =
    resource.kind === "story"
      ? [
          "title",
          "statusId",
          "assigneeId",
          "priority",
          "archivedAt",
          "deletedAt",
        ]
      : ["name", "shortSummary", "statusId", "leadUser", "priority"];
  const changes = Object.fromEntries(
    fields.filter((key) => key in patch).map((key) => [key, patch[key]]),
  );
  return {
    ...resource,
    value: { ...resource.value, ...changes },
  } as NativeLinkResource;
};
const isAccessDenied = (error: unknown) =>
  typeof error === "object" &&
  error !== null &&
  "status" in error &&
  [401, 403, 404].includes(Number(error.status));

export const fetchNativeLink = async (
  target: NativeLinkTarget,
  ctx: WorkspaceCtx,
  options: { signal?: AbortSignal; refresh?: boolean } = {},
): Promise<NativeLinkResource | null> => {
  if (target.kind === "unsupported") return null;
  const params = new URLSearchParams();
  if (target.kind === "objective") params.set("teamId", target.teamId);
  if (options.refresh) params.set("refresh", "true");
  const search = params.toString();
  try {
    const response = await get<ApiResponse<NativeLinkResource>>(
      `previews/native/${target.kind}/${encodeURIComponent(target.id)}${search ? `?${search}` : ""}`,
      ctx,
      { signal: options.signal, cache: "no-store", retry: 0 },
    );
    const resource = response.data;
    if (!resource || resource.kind !== target.kind) return null;
    if (resource.kind === "story" && resource.value.deletedAt) return null;
    if (
      resource.kind === "objective" &&
      target.kind === "objective" &&
      resource.value.teamId !== target.teamId
    )
      return null;
    return resource;
  } catch (error) {
    if (isAccessDenied(error)) return null;
    throw error;
  }
};

const resourceKey = (
  workspace: string,
  resource: NativeLinkResource,
): QueryKey => {
  switch (resource.kind) {
    case "story":
      return storyKeys.detail(workspace, resource.value.id);
    case "objective":
      return objectiveKeys.objective(workspace, resource.value.id);
    case "document":
      return documentKeys.detail(workspace, resource.value.id);
  }
};

const observedClients = new WeakSet<QueryClient>();
const observePreviewInvalidations = (client: QueryClient) => {
  if (observedClients.has(client)) return;
  observedClients.add(client);
  // Follow the QueryClient lifetime so navigating away from a card does not
  // leave its durable copy fresh after a later edit elsewhere in the app.
  client.getQueryCache().subscribe((event) => {
    if (event.type !== "updated") return;
    const invalidated = event.action.type === "invalidate";
    const denied =
      event.action.type === "error" && isAccessDenied(event.query.state.error);
    if (!invalidated && !denied) return;
    for (const query of client
      .getQueryCache()
      .findAll({ queryKey: ["native-link-preview"] })) {
      const cached: unknown = query.state.data;
      const workspace = query.queryKey[2];
      if (
        !isCachedPreview(cached) ||
        !cached.resource ||
        typeof workspace !== "string"
      )
        continue;
      if (
        hashKey(resourceKey(workspace, cached.resource)) !==
        event.query.queryHash
      )
        continue;
      removePreviewCache(query.queryHash);
      if (denied) {
        void client.cancelQueries({ queryKey: query.queryKey, exact: true });
        client.setQueryData<CachedPreview>(query.queryKey, {
          ...cached,
          resource: null,
        });
      } else {
        void client.invalidateQueries({
          predicate: (candidate) => candidate.queryHash === query.queryHash,
        });
      }
    }
  });
};

/** Summaries have their own identity-scoped cache; full resources remain owned by their features. */
export const useNativeLink = (
  target: NativeLinkTarget,
  enabled: boolean,
  active = enabled,
) => {
  const { data: session } = useSession();
  const client = useQueryClient();
  const identity = session?.user.id;
  const cacheReady = usePreviewCacheReady();
  const queryKey = [
    "native-link-preview",
    identity,
    target.workspaceSlug,
    target.kind,
    target.kind === "unsupported" ? "" : target.id,
    target.kind === "objective" ? target.teamId : "",
  ];
  const queryHash = hashKey(queryKey);
  const persisted = useMemo(
    () => (identity ? readPreviewCache(queryHash, isCachedPreview) : undefined),
    [identity, queryHash],
  );
  const canFetch =
    enabled && Boolean(identity) && target.kind !== "unsupported";
  const access = useQuery({
    queryKey,
    queryFn: async ({ signal }): Promise<CachedPreview> => {
      const requestedAt = Date.now();
      const refresh = refreshRequests.get(client)?.has(queryHash) ?? false;
      const resource = await fetchNativeLink(
        target,
        { session, workspaceSlug: target.workspaceSlug },
        { signal, refresh },
      );
      const preview = { resource, requestedAt };
      if (!signal.aborted) {
        refreshRequests.get(client)?.delete(queryHash);
        if (resource) writePreviewCache(queryHash, preview, requestedAt);
        else removePreviewCache(queryHash);
      }
      return preview;
    },
    initialData: persisted?.data,
    initialDataUpdatedAt: persisted?.savedAt,
    enabled: canFetch && active,
    staleTime: (query) =>
      query.state.data?.resource ? PREVIEW_CACHE_TTL_MS : 30_000,
    gcTime: PREVIEW_CACHE_TTL_MS,
    retry: (failures, error) => {
      // Explicit refresh failures surface immediately. Initial/background reads
      // get one retry so a brief outage does not require switching tabs.
      if (refreshRequests.get(client)?.delete(queryHash)) return false;
      return failures < 1 && !isAccessDenied(error);
    },
    retryDelay: 500,
    refetchInterval: canFetch && active ? PREVIEW_CACHE_TTL_MS : false,
    refetchIntervalInBackground: false,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  });
  const resource = access.data?.resource;
  const canonicalKey = resource
    ? resourceKey(target.workspaceSlug, resource)
    : ["native-link-preview-unresolved", identity, target.href];
  const canonical = useQuery<NativeLinkResource["value"] | null>({
    queryKey: canonicalKey,
    queryFn: skipToken,
    enabled: false,
    gcTime: PREVIEW_CACHE_TTL_MS,
  });
  const pendingChanges = useMutationState({
    filters: { status: "pending" },
    select: (mutation) => mutation.state.variables,
  });

  useEffect(() => {
    observePreviewInvalidations(client);
  }, [client]);

  useEffect(() => {
    if (!canFetch) return;
    return client.getMutationCache().subscribe((event) => {
      if (event.type !== "updated" || event.action.type !== "success") return;
      const query = client.getQueryCache().get(queryHash);
      if (!query) return;
      client.setQueryData<CachedPreview>(
        query.queryKey,
        (cached) => {
          if (!cached?.resource) return cached;
          const updated = applyWorkMutation(
            cached.resource,
            event.mutation.state.variables,
          );
          if (updated === cached.resource) return cached;
          const preview = { ...cached, resource: updated };
          if (updated.kind === "story" && updated.value.deletedAt)
            removePreviewCache(queryHash);
          else writePreviewCache(queryHash, preview, cached.requestedAt);
          return preview;
        },
        { updatedAt: query.state.dataUpdatedAt },
      );
    });
  }, [canFetch, client, queryHash]);

  const [leaseClock, setLeaseClock] = useState(() => ({
    queryHash,
    now: Date.now(),
  }));
  const requestedAt = access.data?.requestedAt;
  useEffect(() => {
    const now = Date.now();
    setLeaseClock({ queryHash, now });
    if (!requestedAt) return;
    const remaining = requestedAt + PREVIEW_CACHE_TTL_MS - now;
    if (remaining <= 0) return;
    const timer = setTimeout(() => {
      setLeaseClock({ queryHash, now: Date.now() });
    }, remaining);
    return () => {
      clearTimeout(timer);
    };
  }, [queryHash, requestedAt]);

  const canonicalDenied =
    isAccessDenied(canonical.error) &&
    canonical.errorUpdatedAt >= access.dataUpdatedAt;
  useEffect(() => {
    if (!canonicalDenied) return;
    removePreviewCache(queryHash);
    const query = client.getQueryCache().get(queryHash);
    if (query)
      client.setQueryData<CachedPreview>(query.queryKey, (cached) =>
        cached?.resource ? { ...cached, resource: null } : cached,
      );
  }, [canonicalDenied, client, queryHash]);

  // A successful scoped preview remains usable throughout its one-day lease,
  // including remounts/offscreen cards and temporary refresh failures.
  const authorized =
    cacheReady &&
    canFetch &&
    Boolean(resource) &&
    Boolean(
      access.data &&
        leaseClock.queryHash === queryHash &&
        leaseClock.now - access.data.requestedAt < PREVIEW_CACHE_TTL_MS,
    ) &&
    !isAccessDenied(access.error) &&
    !canonicalDenied;
  const canonicalIsNewer =
    canonical.data &&
    canonical.dataUpdatedAt >= (access.data?.requestedAt ?? Infinity);
  const base =
    authorized && resource
      ? ({
          ...resource,
          value: canonicalIsNewer ? canonical.data : resource.value,
        } as NativeLinkResource)
      : null;
  const data = base
    ? pendingChanges.reduce<NativeLinkResource>(applyWorkMutation, base)
    : null;
  // An expired lease still waits for a fresh access check before showing data.
  const awaitingAccess =
    canFetch &&
    !data &&
    (access.isFetching || (!active && Boolean(resource) && !access.isError));
  const loading =
    target.kind !== "unsupported" &&
    (!enabled || (Boolean(identity) && access.isPending) || awaitingAccess);
  const refresh = async () => {
    if (!canFetch) return;
    requestRefresh(client, queryHash);
    const result = await access.refetch({
      cancelRefetch: true,
      throwOnError: true,
    });
    if (!result.data?.resource)
      throw new Error("This preview is no longer available.");
  };
  return {
    data: data?.kind === "story" && data.value.deletedAt ? null : data,
    loading,
    unavailable:
      target.kind !== "unsupported" &&
      enabled &&
      !loading &&
      (!identity || (!access.isPending && !data)),
    retry: () => {
      requestRefresh(client, queryHash);
      void access.refetch();
    },
    refresh,
    refreshing: Boolean(access.isFetching && access.data),
    refreshedAt: resource?.refreshedAt
      ? Date.parse(resource.refreshedAt)
      : undefined,
    refreshError: access.error,
  };
};
