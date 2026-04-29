import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { userKeys } from "@/constants/keys";
import {
  createGitHubUserLinkSessionAction,
  linkGitHubUserAction,
  unlinkGitHubUserAction,
} from "@/lib/actions/github/link-user";

export const useCreateGitHubUserLinkSession = () =>
  useMutation({
    mutationFn: (returnTo: string) =>
      createGitHubUserLinkSessionAction(returnTo),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
      }
    },
  });

export const useLinkGitHubUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { code: string; state: string }) =>
      linkGitHubUserAction(input),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      toast.success("GitHub", {
        description: "Your GitHub account has been linked.",
      });
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },
  });
};

export const useUnlinkGitHubUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => unlinkGitHubUserAction(),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      toast.success("GitHub", {
        description: "Your GitHub account has been unlinked.",
      });
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },
  });
};
