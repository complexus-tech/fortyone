import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { teamKeys } from "@/constants/keys";
import { addTeamMember, removeTeamMember } from "../actions";

export const useTeamMemberMutations = (teamId: string) => {
  const queryClient = useQueryClient();

  const addMember = useMutation({
    mutationFn: ({
      userId,
      role,
    }: {
      userId: string;
      role: "admin" | "member";
    }) => addTeamMember(teamId, userId, role),
    onSuccess: () => {
      toast.success("Success", {
        description: "Team member added successfully",
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });

  const removeMember = useMutation({
    mutationFn: (userId: string) => removeTeamMember(teamId, userId),
    onSuccess: () => {
      toast.success("Success", {
        description: "Team member removed successfully",
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });

  return { addMember, removeMember };
};
