import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { userKeys } from "@/constants/keys";
import { updateProfile } from "../actions/update-profile";
import type { User } from "@/types";
import type { UpdateProfile } from "../actions/update-profile";

export const useUpdateProfileMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (updates: UpdateProfile) => updateProfile(updates),

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: userKeys.profile() });
      const previousProfile = queryClient.getQueryData<User>(
        userKeys.profile()
      );

      if (previousProfile) {
        queryClient.setQueryData<User>(userKeys.profile(), {
          ...previousProfile,
          ...updates,
        });
      }

      return { previousProfile };
    },
    onSuccess: () => {
      toast.success("Profile updated");
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },

    onError: (error, variables, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(userKeys.profile(), context.previousProfile);
      }

      toast.error("Failed to update profile", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
        },
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },
  });

  return mutation;
};
