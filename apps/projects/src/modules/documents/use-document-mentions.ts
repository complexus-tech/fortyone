"use client";

import { useCallback, useEffect } from "react";
import type { Editor } from "@tiptap/core";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import { documentKeys } from "@/shared/documents/keys";
import { searchQuery } from "@/modules/search/public/search";
import type { DocumentMentionItem } from "./document-mention-list";
import { setDocumentMentionSearch } from "./document-mentions";

export type DocumentMentionRecipients = {
  members: {
    id: string;
    name: string;
    username: string;
    avatarUrl?: string | null;
  }[];
  mayaEnabled: boolean;
  mayaId?: string;
};

export const useDocumentMentionRecipients = (documentId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: [
      ...documentKeys.detail(workspaceSlug, documentId),
      "mention-recipients",
      session?.user.id,
    ],
    queryFn: async () => {
      const response = await get<ApiResponse<DocumentMentionRecipients>>(
        `documents/${documentId}/comment-mentions`,
        { session: session!, workspaceSlug },
      );
      if (!response.data) throw new Error("Could not load people");
      return response.data;
    },
    enabled: Boolean(session && documentId),
    staleTime: 30_000,
  });
};

export function useDocumentMentions(
  editor: Editor | null,
  documentId: string,
  surface: "body" | "comment",
) {
  const recipients = useDocumentMentionRecipients(documentId);
  const { data: recipientData, refetch: refetchRecipients } = recipients;
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const search = useCallback(
    async (query: string): Promise<DocumentMentionItem[]> => {
      const data = recipientData ?? (await refetchRecipients()).data;
      if (!data)
        return [
          {
            kind: "message",
            id: "error",
            label: "Couldn’t load people. Try typing again.",
          },
        ];
      const normalized = query.trim().toLowerCase();
      const people: DocumentMentionItem[] = data.members
        .filter((member) =>
          `${member.name} ${member.username}`
            .toLowerCase()
            .includes(normalized),
        )
        .slice(0, 6)
        .map((member) => ({
          id: member.id,
          label: member.name || member.username,
          avatar: member.avatarUrl,
          kind: "user",
        }));
      if (
        surface === "comment" &&
        data.mayaEnabled &&
        data.mayaId &&
        "maya".includes(normalized)
      )
        people.unshift({ id: data.mayaId, label: "Maya", kind: "agent" });
      if (surface !== "body" || !session || !normalized) return people;
      try {
        const results = await queryClient.fetchQuery({
          queryKey: [
            ...documentKeys.detail(workspaceSlug, documentId),
            "reference-search",
            session.user.id,
            normalized,
          ],
          queryFn: () =>
            searchQuery(
              { session, workspaceSlug },
              { query: normalized, type: "all", pageSize: 8 },
            ),
          staleTime: 15_000,
        });
        return [
          ...people.slice(0, 4),
          ...results.stories.slice(0, 4).map(
            (story): DocumentMentionItem => ({
              kind: "story",
              id: story.id,
              label: story.title,
              displayId: `${story.teamCode || story.team?.code}-${story.sequenceId}`,
            }),
          ),
          ...results.objectives.slice(0, 4).map(
            (objective): DocumentMentionItem => ({
              kind: "objective",
              id: objective.id,
              label: objective.name,
            }),
          ),
        ];
      } catch {
        return people.length
          ? people
          : [
              {
                kind: "message",
                id: "error",
                label: "Couldn’t load references. Try typing again.",
              },
            ];
      }
    },
    [
      documentId,
      queryClient,
      recipientData,
      refetchRecipients,
      session,
      surface,
      workspaceSlug,
    ],
  );
  useEffect(() => {
    if (editor) setDocumentMentionSearch(editor, search);
  }, [editor, search]);
  return recipients;
}
