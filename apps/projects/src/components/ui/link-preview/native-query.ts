"use client";

import type { QueryClient, QueryKey } from "@tanstack/react-query";
import {
  hashKey,
  skipToken,
  useQuery,
  useQueryClient,
  useMutationState,
} from "@tanstack/react-query";
import { useEffect, useState } from "react";
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

const ACCESS_LEASE_MS = 30_000;
// A query can have several card observers. The refresh signal must belong to
// the shared query, not the particular card whose menu was clicked.
const refreshRequests = new WeakMap<QueryClient, Set<string>>();
// QueryCache passes the same event to every card observer. Each affected
// summary needs one refresh per event; a later edit must still start a new one.
const handledInvalidations = new WeakMap<object, Set<string>>();
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

/** Summaries have their own identity-scoped cache; full resources remain owned by their features. */
export const useNativeLink = (
  target: NativeLinkTarget,
  enabled: boolean,
  active = enabled,
) => {
  const { data: session } = useSession();
  const client = useQueryClient();
  const identity = session?.user.id;
  const queryKey = [
    "native-link-preview",
    identity,
    target.workspaceSlug,
    target.kind,
    target.kind === "unsupported" ? "" : target.id,
    target.kind === "objective" ? target.teamId : "",
  ];
  const queryHash = hashKey(queryKey);
  const [accessScope, setAccessScope] = useState(() => ({
    queryHash,
    mountedAt: Date.now(),
  }));
  useEffect(() => {
    const mountedAt = Date.now();
    setAccessScope((current) =>
      current.queryHash === queryHash ? current : { queryHash, mountedAt },
    );
  }, [queryHash]);
  const canFetch =
    enabled && Boolean(identity) && target.kind !== "unsupported";
  const access = useQuery({
    queryKey,
    queryFn: async ({ signal }): Promise<CachedPreview> => {
      const requestedAt = Date.now();
      const refresh = refreshRequests.get(client)?.delete(queryHash) ?? false;
      const resource = await fetchNativeLink(
        target,
        { session, workspaceSlug: target.workspaceSlug },
        { signal, refresh },
      );
      return { resource, requestedAt };
    },
    enabled: canFetch && active,
    staleTime: ACCESS_LEASE_MS,
    gcTime: 5 * 60_000,
    retry: false,
    refetchInterval: canFetch && active ? ACCESS_LEASE_MS : false,
    refetchIntervalInBackground: false,
    refetchOnWindowFocus: true,
    refetchOnReconnect: "always",
  });
  const resource = access.data?.resource;
  const canonicalKey = resource
    ? resourceKey(target.workspaceSlug, resource)
    : ["native-link-preview-unresolved", identity, target.href];
  const canonicalHash = hashKey(canonicalKey);
  const canonical = useQuery<NativeLinkResource["value"] | null>({
    queryKey: canonicalKey,
    queryFn: skipToken,
    enabled: false,
  });
  const pendingChanges = useMutationState({
    filters: { status: "pending" },
    select: (mutation) => mutation.state.variables,
  });

  useEffect(() => {
    if (!canFetch || !resource) return;
    return client.getQueryCache().subscribe((event) => {
      // Existing mutations and realtime reconciliation invalidate canonical
      // keys. Translate those signals to a small preview request, never a full
      // document/story fetch owned by this component.
      if (
        event.query.queryHash === canonicalHash &&
        event.type === "updated" &&
        event.action.type === "invalidate"
      ) {
        let refreshedQueries = handledInvalidations.get(event);
        if (!refreshedQueries) {
          refreshedQueries = new Set();
          handledInvalidations.set(event, refreshedQueries);
        }
        if (refreshedQueries.has(queryHash)) return;
        refreshedQueries.add(queryHash);
        void client.invalidateQueries({
          predicate: (query) => query.queryHash === queryHash,
        });
      }
    });
  }, [canFetch, client, canonicalHash, queryHash, resource]);

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
          return updated === cached.resource
            ? cached
            : { ...cached, resource: updated };
        },
        { updatedAt: query.state.dataUpdatedAt },
      );
    });
  }, [canFetch, client, queryHash]);

  const authorized =
    canFetch &&
    Boolean(resource) &&
    accessScope.queryHash === queryHash &&
    Boolean(
      access.data &&
        (access.dataUpdatedAt >= accessScope.mountedAt ||
          (active && !access.isStale)),
    ) &&
    !isAccessDenied(access.error) &&
    !(
      isAccessDenied(canonical.error) &&
      canonical.errorUpdatedAt >= access.dataUpdatedAt
    );
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
  // A remounted cached query can already be in the success state while its
  // current access check is still pending. Keep stale fields hidden, but do not
  // describe that normal revalidation/visibility wait as an access failure.
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
