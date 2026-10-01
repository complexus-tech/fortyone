import { createApiClient } from "api-client";
import { z } from "zod";
import { getApiUrl } from "@/lib/api-url";

const statusSchema = z.object({
  data: z.object({ enabled: z.boolean(), requireSSO: z.boolean() }),
});

export const getWorkspaceSSOStatus = async (workspaceSlug: string) => {
  const response = await createApiClient(getApiUrl()).get(
    `auth/sso/${encodeURIComponent(workspaceSlug)}/status`,
    { credentials: "omit", retry: 0 },
  );
  return statusSchema.parse(await response.json()).data;
};
