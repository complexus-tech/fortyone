import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { teamKeys } from "@/constants/keys";
import type { Team } from "../types";
import { createTeam } from "../actions";

export const useCreateTeamMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: createTeam,
    onMutate: (data) => {
      const previousTeams = queryClient.getQueryData<Team[]>(teamKeys.lists());

      const newTeam: Team = {
        ...data,
        id: "1",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        workspaceId: "1",
        members: [],
      };

      queryClient.setQueryData(teamKeys.lists(), (old: Team[]) => [
        ...old,
        newTeam,
      ]);

      return { previousTeams };
    },
    onSuccess: (data) => {
      toast.success("Success", {
        description: "Team created successfully",
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      router.push(`/settings/workspace/teams/${data.data?.id}`);
    },
    onError: (error, _, context) => {
      queryClient.setQueryData(teamKeys.lists(), context?.previousTeams);
      toast.error("Failed to create team", {
        description: error.message || "Team code must be unique",
      });
    },
  });
};
