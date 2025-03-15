import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { terminologyKeys } from "@/constants/keys";
import type { Terminology } from "@/types";
import type { UpdateTerminology } from "../../actions/terminology/update";
import { updateTerminologyAction } from "../../actions/terminology/update";

export const useUpdateTerminologyMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (payload: UpdateTerminology) =>
      updateTerminologyAction(payload),

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: terminologyKeys.all });
      const previousTerminology = queryClient.getQueryData<Terminology>(
        terminologyKeys.all,
      );

      if (previousTerminology) {
        queryClient.setQueryData<Terminology>(terminologyKeys.all, {
          ...previousTerminology,
          ...updates,
        });
      }

      return { previousTerminology };
    },

    onError: (error, variables, context) => {
      if (context?.previousTerminology) {
        queryClient.setQueryData(
          terminologyKeys.all,
          context.previousTerminology,
        );
      }
      toast.error("Failed to update terminology", {
        description: error.message || "Your changes were not saved",
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
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: terminologyKeys.all });
    },
  });

  return mutation;
};
