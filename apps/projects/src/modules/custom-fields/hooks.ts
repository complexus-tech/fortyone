"use client";

import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { analyticsKeys } from "@/constants/keys";
import { createStoryValueBatches } from "./value-utils";
import {
  archiveCustomField,
  buildCustomFieldReport,
  createCustomField,
  getStoryCustomFields,
  getTeamCustomFields,
  updateCustomField,
  updateStoryCustomFields,
  getCustomFieldStoryValues,
} from "./api";
import type {
  CustomFieldDraft,
  CustomFieldReportInput,
  CustomFieldUpdate,
  CustomFieldValue,
  CustomFieldStoryValues,
} from "./types";

export const customFieldKeys = {
  all: (workspace: string) => ["custom-fields", workspace] as const,
  team: (workspace: string, teamId: string) =>
    ["custom-fields", workspace, "team", teamId] as const,
  story: (workspace: string, storyId: string) =>
    ["custom-fields", workspace, "story", storyId] as const,
  batches: (workspace: string) =>
    ["custom-fields", workspace, "story-values"] as const,
};

const useCustomFieldContext = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return { session, workspaceSlug };
};

export const useTeamCustomFields = (teamId?: string) => {
  const ctx = useCustomFieldContext();
  return useQuery({
    queryKey: customFieldKeys.team(ctx.workspaceSlug, teamId ?? ""),
    queryFn: () => getTeamCustomFields(teamId!, ctx),
    enabled: Boolean(ctx.session && ctx.workspaceSlug && teamId),
    staleTime: 60_000,
  });
};

export const useCustomFieldMutations = (teamId: string) => {
  const ctx = useCustomFieldContext();
  const queryClient = useQueryClient();
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: customFieldKeys.all(ctx.workspaceSlug),
    });
  const create = useMutation({
    mutationFn: (input: CustomFieldDraft) =>
      createCustomField(teamId, input, ctx),
    onSuccess: invalidate,
  });
  const update = useMutation({
    mutationFn: ({
      fieldId,
      input,
    }: {
      fieldId: string;
      input: CustomFieldUpdate;
    }) => updateCustomField(teamId, fieldId, input, ctx),
    onSuccess: invalidate,
  });
  const archive = useMutation({
    mutationFn: (fieldId: string) => archiveCustomField(teamId, fieldId, ctx),
    onSuccess: invalidate,
  });
  return { create, update, archive };
};

export const useStoryCustomFields = (storyId: string) => {
  const ctx = useCustomFieldContext();
  return useQuery({
    queryKey: customFieldKeys.story(ctx.workspaceSlug, storyId),
    queryFn: () => getStoryCustomFields(storyId, ctx),
    enabled: Boolean(ctx.session && ctx.workspaceSlug && storyId),
    staleTime: 30_000,
  });
};

export const useUpdateStoryCustomFields = (storyId: string) => {
  const ctx = useCustomFieldContext();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      values,
      expectedVersion,
    }: {
      values: CustomFieldValue[];
      expectedVersion?: number;
    }) => updateStoryCustomFields(storyId, values, expectedVersion, ctx),
    onSuccess: async (snapshot) => {
      queryClient.setQueryData(
        customFieldKeys.story(ctx.workspaceSlug, storyId),
        snapshot,
      );
      queryClient.setQueriesData<CustomFieldStoryValues>(
        { queryKey: customFieldKeys.batches(ctx.workspaceSlug) },
        (current) =>
          current && {
            ...current,
            items: current.items.map((item) =>
              item.storyId === storyId
                ? {
                    ...item,
                    values: snapshot.values,
                    version: snapshot.version ?? item.version,
                  }
                : item,
            ),
          },
      );
      await queryClient.invalidateQueries({
        queryKey: customFieldKeys.batches(ctx.workspaceSlug),
      });
      await queryClient.invalidateQueries({
        queryKey: analyticsKeys.all(ctx.workspaceSlug),
      });
    },
  });
};

export const useCustomFieldReport = () => {
  const ctx = useCustomFieldContext();
  return useMutation({
    mutationFn: (input: CustomFieldReportInput) =>
      buildCustomFieldReport(input, ctx),
  });
};

export const useCustomFieldStoryValues = (
  storyIds: string[],
  enabled = true,
) => {
  const ctx = useCustomFieldContext();
  const queryClient = useQueryClient();
  const batches = createStoryValueBatches(storyIds);
  const queries = useQueries({
    queries: batches.map((ids) => ({
      queryKey: [...customFieldKeys.batches(ctx.workspaceSlug), ids],
      queryFn: () => getCustomFieldStoryValues(ids, ctx),
      enabled: Boolean(enabled && ctx.session && ctx.workspaceSlug),
      staleTime: 30_000,
    })),
  });
  return {
    items: new Map(
      queries
        .flatMap((query) => query.data?.items ?? [])
        .map((item) => [item.storyId, item]),
    ),
    isPending: queries.some((query) => query.isPending),
    isError: queries.some((query) => query.isError),
    refetchStory: async (storyId: string) => {
      const index = batches.findIndex((ids) => ids.includes(storyId));
      if (index < 0) {
        throw new Error("This item's custom fields are unavailable.");
      }
      const query = queries[index];
      const result = await query.refetch({ throwOnError: true });
      const item = result.data?.items.find(
        (candidate) => candidate.storyId === storyId,
      );
      if (!item) throw new Error("Fields could not be reloaded.");
      await queryClient.invalidateQueries({
        queryKey: customFieldKeys.story(ctx.workspaceSlug, storyId),
      });
      return item;
    },
  };
};

export const useVisibleTeamCustomFields = (
  teamIds: string[],
  enabled = true,
) => {
  const ctx = useCustomFieldContext();
  return useQueries({
    queries: Array.from(new Set(teamIds))
      .sort()
      .map((teamId) => ({
        queryKey: customFieldKeys.team(ctx.workspaceSlug, teamId),
        queryFn: () => getTeamCustomFields(teamId, ctx),
        enabled: Boolean(enabled && ctx.session && ctx.workspaceSlug),
        staleTime: 60_000,
      })),
  });
};
