"use client";
import { useQuery } from "@tanstack/react-query";
import { Box, Button, Text } from "ui";
import { z } from "zod";
import { useSession } from "@/lib/auth/client";
import { get } from "@/lib/http";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { ImportDraft } from "../schema";

const receiptPageSchema = z.object({
  data: z.object({
    items: z.array(
      z.object({
        sourceKey: z.string(),
        storyId: z.uuid().nullable(),
        errorCode: z.string().nullable(),
        errorMessage: z.string().nullable(),
      }),
    ),
    hasMore: z.boolean(),
  }),
});

export const ImportResumePanel = ({ draft }: { draft: ImportDraft }) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const query = useQuery({
    queryKey: [
      "import-receipts",
      workspaceSlug,
      draft.fileHash,
      draft.sourceNamespace,
    ],
    enabled: Boolean(session?.token),
    queryFn: async () => {
      const items: z.infer<typeof receiptPageSchema>["data"]["items"] = [];
      const provider = draft.sourceType === "jira_csv" ? "jira_csv" : "file";
      for (let offset = 0; offset <= 20_000; offset += 500) {
        const search = new URLSearchParams({
          provider,
          sourceDigest: draft.fileHash,
          offset: String(offset),
        });
        if (draft.sourceNamespace)
          search.set("sourceNamespace", draft.sourceNamespace);
        // eslint-disable-next-line no-await-in-loop -- Receipt pages are bounded and fetched only after the previous continuation is known.
        const page = await get(
          `stories/import/receipts?${search.toString()}`,
          { session, workspaceSlug },
          undefined,
          (response) => receiptPageSchema.parse(response).data,
        );
        items.push(...page.items);
        if (!page.hasMore) break;
      }
      return items;
    },
    retry: false,
  });
  if (query.isPending) return null;
  if (query.isError)
    return (
      <Box className="border-border mt-5 rounded-xl border p-4">
        <Text color="muted">
          Previous import receipts could not be loaded. Stable source IDs still
          protect task retries.
        </Text>
        <Button
          className="mt-3"
          color="tertiary"
          onClick={() => {
            void query.refetch();
          }}
        >
          Retry receipt check
        </Button>
      </Box>
    );
  if (!query.data.length) return null;
  const completed = query.data.filter(
    (item) => item.storyId && !item.errorCode,
  ).length;
  return (
    <Box className="bg-surface-muted mt-5 rounded-xl p-4">
      <Text className="font-medium">Continue a previous import</Text>
      <Text className="mt-1 leading-6" color="muted">
        {completed} source tasks are saved. Continue this reviewed import to
        finish comments and links while retaining changes made in FortyOne.
      </Text>
    </Box>
  );
};
