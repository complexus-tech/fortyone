import type { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { get, post, put, remove } from "@/lib/http";
import {
  customFieldSchema,
  customFieldReportSchema,
  storyCustomFieldsSchema,
  customFieldStoryValuesSchema,
} from "./types";
import type {
  CustomFieldDraft,
  CustomFieldUpdate,
  CustomFieldValue,
  CustomFieldReportInput,
} from "./types";

const decodeData =
  <T>(schema: z.ZodType<T>) =>
  (response: unknown): T => {
    if (
      typeof response !== "object" ||
      response === null ||
      !("data" in response)
    )
      throw new Error("Missing API data");
    return schema.parse(response.data);
  };

export const getTeamCustomFields = (teamId: string, ctx: WorkspaceCtx) =>
  get(
    `teams/${teamId}/custom-fields`,
    ctx,
    undefined,
    decodeData(customFieldSchema.array()),
  );

export const createCustomField = (
  teamId: string,
  input: CustomFieldDraft,
  ctx: WorkspaceCtx,
) =>
  post(
    `teams/${teamId}/custom-fields`,
    input,
    ctx,
    undefined,
    decodeData(customFieldSchema),
  );

export const updateCustomField = (
  teamId: string,
  fieldId: string,
  input: CustomFieldUpdate,
  ctx: WorkspaceCtx,
) =>
  put(
    `teams/${teamId}/custom-fields/${fieldId}`,
    input,
    ctx,
    undefined,
    decodeData(customFieldSchema),
  );

export const archiveCustomField = (
  teamId: string,
  fieldId: string,
  ctx: WorkspaceCtx,
) => remove(`teams/${teamId}/custom-fields/${fieldId}`, ctx);

export const getStoryCustomFields = (storyId: string, ctx: WorkspaceCtx) =>
  get(
    `stories/${storyId}/custom-fields`,
    ctx,
    undefined,
    decodeData(storyCustomFieldsSchema),
  );

export const updateStoryCustomFields = (
  storyId: string,
  values: CustomFieldValue[],
  expectedVersion: number | undefined,
  ctx: WorkspaceCtx,
) =>
  put(
    `stories/${storyId}/custom-fields`,
    { values, ...(expectedVersion === undefined ? {} : { expectedVersion }) },
    ctx,
    undefined,
    decodeData(storyCustomFieldsSchema),
  );

export const buildCustomFieldReport = (
  input: CustomFieldReportInput,
  ctx: WorkspaceCtx,
) =>
  post(
    "analytics/custom-field-report",
    input,
    ctx,
    undefined,
    decodeData(customFieldReportSchema),
  );

export const getCustomFieldStoryValues = (
  storyIds: string[],
  ctx: WorkspaceCtx,
) =>
  post(
    "custom-fields/story-values",
    { storyIds },
    ctx,
    undefined,
    decodeData(customFieldStoryValuesSchema),
  );
