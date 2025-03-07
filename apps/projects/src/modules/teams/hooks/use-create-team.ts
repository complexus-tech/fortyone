import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { statusKeys, teamKeys } from "@/constants/keys";
import type { Team } from "../types";
import { createTeam } from "../actions";

export const useCreateTeamMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const toastId = "create-team";

  const mutation = useMutation({
    mutationFn: createTeam,
    onMutate: (data) => {
      const previousTeams = queryClient.getQueryData<Team[]>(teamKeys.lists());

      const newTeam: Team = {
        ...data,
        id: "1",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        workspaceId: "1",
        memberCount: 1,
      };

      queryClient.setQueryData(teamKeys.lists(), (old: Team[]) => [
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
      queryClient.setQueryData(teamKeys.lists(), context?.previousTeams);
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
    onSuccess: (data) => {
      toast.success("Success", {
        description: "Team created successfully",
        id: toastId,
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      queryClient.invalidateQueries({ queryKey: statusKeys.lists() });
      router.push(`/settings/workspace/teams/${data.data?.id}`);
    },
  });

  return mutation;
};
