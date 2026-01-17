import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { statusKeys, teamKeys } from "@/constants/keys";
import type { Team } from "../types";
import { createTeam } from "../actions";

export const useCreateTeamMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { analytics } = useAnalytics();
  const { workspaceSlug } = useWorkspacePath();
  const toastId = "create-team";

  const mutation = useMutation({
    mutationFn: (data: Parameters<typeof createTeam>[0]) => createTeam(data, workspaceSlug),
    onMutate: (data) => {
      const previousTeams = queryClient.getQueryData<Team[]>(teamKeys.lists(workspaceSlug));

      const newTeam: Team = {
        ...data,
        id: "1",
        sprintsEnabled: false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        workspaceId: "1",
        memberCount: 1,
      };

      queryClient.setQueryData(teamKeys.lists(workspaceSlug), (old: Team[]) => [
        ...old,
        newTeam,
      ]);

      toast.loading("Creating team...", {
        description: "Please wait while we create the team",
        id: toastId,
      });

      return { previousTeams };
    },

    onError: (error, variables, context) => {
      queryClient.setQueryData(teamKeys.lists(workspaceSlug), context?.previousTeams);
      toast.error("Failed to create team", {
        description: error.message || "Team code must be unique",
        id: toastId,
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Failed to create team", {
          description: JSON.stringify(res.error),
          id: toastId,
        });
        throw new Error(res.error.message);
      }

      const data = res.data;

      // Track team creation
      if (data) {
        analytics.track("team_created", {
          teamId: data.id,
          name: data.name,
          code: data.code,
          isPrivate: data.isPrivate,
        });
      }

      toast.success("Success", {
        description: "Team created successfully",
        id: toastId,
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists(workspaceSlug) });
      queryClient.invalidateQueries({ queryKey: statusKeys.lists(workspaceSlug) });
      router.push(`/settings/workspace/teams/${data?.id}`);
    },
  });

  return mutation;
};
