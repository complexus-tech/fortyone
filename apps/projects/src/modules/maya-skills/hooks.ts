import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { ApiError } from "@/lib/http";
import {
  createMayaSkill,
  deleteMayaSkill,
  listMayaSkills,
  updateMayaSkill,
} from "./api";
import type { MayaSkillInput, MayaSkillUpdate } from "./types";

export const mayaSkillsKey = (workspaceSlug: string, userId: string) =>
  ["maya-skills", workspaceSlug, userId] as const;

const useSkillScope = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return {
    ctx: { session, workspaceSlug },
    key: mayaSkillsKey(workspaceSlug, session?.user.id ?? ""),
    enabled: Boolean(session?.user.id && workspaceSlug),
  };
};

export const useMayaSkills = (enabled = true) => {
  const scope = useSkillScope();
  return useQuery({
    queryKey: scope.key,
    queryFn: () => listMayaSkills(scope.ctx),
    enabled: scope.enabled && enabled,
    staleTime: 60_000,
  });
};

export const useMayaSkillMutations = () => {
  const scope = useSkillScope();
  const queryClient = useQueryClient();
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: scope.key });
  return {
    create: useMutation({
      mutationFn: (input: MayaSkillInput) => createMayaSkill(input, scope.ctx),
      onSuccess: invalidate,
    }),
    update: useMutation({
      mutationFn: ({ id, input }: { id: string; input: MayaSkillUpdate }) =>
        updateMayaSkill(id, input, scope.ctx),
      onSuccess: invalidate,
      onError: async (error) => {
        if (error instanceof ApiError && error.status === 409) {
          await invalidate();
        }
      },
    }),
    remove: useMutation({
      mutationFn: (id: string) => deleteMayaSkill(id, scope.ctx),
      onSuccess: invalidate,
    }),
  };
};
