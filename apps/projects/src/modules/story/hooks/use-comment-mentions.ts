import type { Editor } from "@tiptap/core";
import { useEffect } from "react";
import { useMayaAssignee } from "@/lib/hooks/members";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { setStoryCommentMentionUsers } from "../components/story-comment-editor";
import type { MentionItem } from "../components/mentions/list";
import { useStoryMayaRuns } from "./comment-maya";

export const useCommentMentions = (
  editor: Editor | null,
  storyId: string,
  teamId: string,
  allowMaya: boolean,
) => {
  const { data: members } = useTeamMembers(teamId);
  const capability = useStoryMayaRuns(storyId);
  const enabled = allowMaya && capability.data?.enabled === true;
  const { data: maya } = useMayaAssignee(enabled);
  useEffect(() => {
    if (!editor) return;
    const people: MentionItem[] = (members ?? [])
      .filter((member) => !member.isSystem)
      .map((member) => ({
        id: member.id,
        label: member.fullName || member.username,
        username: member.username,
        avatar: member.avatarUrl,
      }));
    if (enabled && maya?.isSystem) {
      people.unshift({
        id: maya.id,
        label: "Maya",
        username: "maya",
        avatar: maya.avatarUrl,
        mentionKind: "agent",
        agentKey: "maya",
      });
    }
    setStoryCommentMentionUsers(editor, people);
  }, [editor, enabled, maya, members]);
  return {
    maya: enabled && maya?.isSystem ? maya : undefined,
    enabled,
    isUnavailable: capability.isError,
  };
};
