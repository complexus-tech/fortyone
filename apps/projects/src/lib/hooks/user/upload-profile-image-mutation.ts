import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { uploadProfileImageAction } from "@/lib/actions/users/upload-profile-image";
import { userKeys } from "@/constants/keys";
import type { User } from "@/types";

export const useUploadProfileImageMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (file: File) => uploadProfileImageAction(file),
    onMutate: async (file) => {
      await queryClient.cancelQueries({ queryKey: userKeys.profile() });
      const previousProfile = queryClient.getQueryData<User>(
        userKeys.profile(),
      );

      if (previousProfile) {
        const tempUrl = URL.createObjectURL(file);
        queryClient.setQueryData<User>(userKeys.profile(), {
          ...previousProfile,
          avatarUrl: tempUrl,
        });
        return { previousProfile, tempUrl };
      }
    },
    onError: (error, variables, context) => {
      if (context?.tempUrl) {
        URL.revokeObjectURL(context.tempUrl);
      }
      if (context?.previousProfile) {
        queryClient.setQueryData(userKeys.profile(), context.previousProfile);
      }
      toast.error("Failed to update profile image", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res, _, context) => {
      if (context.tempUrl) {
        URL.revokeObjectURL(context.tempUrl);
      }
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },
  });

  return mutation;
};
