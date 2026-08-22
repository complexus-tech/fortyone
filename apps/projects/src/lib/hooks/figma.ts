import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { figmaKeys, linkKeys } from "@/constants/keys";
import {
  getFigmaHandoffStatuses,
  getFigmaIntegration,
  getStoryFigmaLinks,
} from "@/lib/queries/figma";
import {
  createFigmaInstallSessionAction,
  deleteFigmaStoryLinkAction,
  disconnectFigmaAction,
  linkFigmaStoryAction,
  refreshFigmaStoryLinkAction,
  resolveFigmaLinkAction,
} from "@/lib/actions/figma";
import { figmaDescriptionSchema } from "@/modules/settings/workspace/integrations/figma/description";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";

export const useFigmaIntegration = ({ enabled = true } = {}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: enabled && Boolean(session && workspaceSlug),
    queryKey: figmaKeys.integration(workspaceSlug),
    queryFn: () => getFigmaIntegration({ session: session!, workspaceSlug }),
  });
};

export const useStoryFigmaLinks = (storyId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: Boolean(session && workspaceSlug && storyId),
    queryKey: figmaKeys.storyLinks(workspaceSlug, storyId),
    queryFn: () =>
      getStoryFigmaLinks(storyId, { session: session!, workspaceSlug }),
    refetchInterval: 60_000,
  });
};

export const useFigmaHandoffStatuses = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    enabled: Boolean(session && workspaceSlug),
    queryKey: figmaKeys.handoffStatuses(workspaceSlug),
    queryFn: () =>
      getFigmaHandoffStatuses({ session: session!, workspaceSlug }),
    refetchInterval: 60_000,
    staleTime: 60_000,
  });
};

export const useCreateFigmaInstallSession = () => {
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async () => {
      const response = await createFigmaInstallSessionAction(workspaceSlug);
      if (response.error?.message || !response.data?.authorizationUrl) {
        throw new Error(
          response.error?.message ??
            "Figma did not return an authorization URL.",
        );
      }
      return response.data;
    },
    onSuccess: ({ authorizationUrl }) => {
      window.location.assign(authorizationUrl);
    },
    onError: (error) => toast.error("Figma", { description: error.message }),
  });
};

export const useDisconnectFigma = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async () => {
      const response = await disconnectFigmaAction(workspaceSlug);
      if (response.error?.message) throw new Error(response.error.message);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: figmaKeys.integration(workspaceSlug),
      });
      toast.success("Figma disconnected");
    },
    onError: (error) => toast.error("Figma", { description: error.message }),
  });
};

export const useResolveFigmaLink = () => {
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async (url: string) => {
      const response = await resolveFigmaLinkAction(workspaceSlug, url);
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message ?? "Figma did not return design metadata.",
        );
      }
      return response.data;
    },
  });
};

export const useExtractFigmaDescription = () =>
  useMutation({
    mutationFn: async (artifact: FigmaArtifact) => {
      if (!artifact.textContent?.length) {
        throw new Error("This design does not contain extractable text.");
      }
      const response = await fetch("/api/figma-description", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          fileName: artifact.fileName,
          nodeName: artifact.nodeName,
          nodeType: artifact.nodeType,
          textContent: artifact.textContent,
        }),
      });
      if (!response.ok) {
        throw new Error(
          (await response.text()) || "The description could not be extracted.",
        );
      }
      const parsed = figmaDescriptionSchema.safeParse(await response.json());
      if (!parsed.success) {
        throw new Error("The AI returned an invalid story description.");
      }
      return parsed.data;
    },
  });

export const useLinkFigmaStory = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async ({ storyId, url }: { storyId: string; url: string }) => {
      const response = await linkFigmaStoryAction(workspaceSlug, storyId, url);
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message ?? "The Figma design could not be linked.",
        );
      }
      return response.data;
    },
    onSuccess: async (_, { storyId }) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: figmaKeys.storyLinks(workspaceSlug, storyId),
        }),
        queryClient.invalidateQueries({ queryKey: linkKeys.story(storyId) }),
        queryClient.invalidateQueries({
          queryKey: figmaKeys.handoffStatuses(workspaceSlug),
        }),
      ]);
    },
  });
};

export const useDeleteFigmaStoryLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async ({
      storyId,
      linkId,
    }: {
      storyId: string;
      linkId: string;
    }) => {
      const response = await deleteFigmaStoryLinkAction(
        workspaceSlug,
        storyId,
        linkId,
      );
      if (response.error?.message) throw new Error(response.error.message);
    },
    onSuccess: async (_, { storyId }) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: figmaKeys.storyLinks(workspaceSlug, storyId),
        }),
        queryClient.invalidateQueries({ queryKey: linkKeys.story(storyId) }),
        queryClient.invalidateQueries({
          queryKey: figmaKeys.handoffStatuses(workspaceSlug),
        }),
      ]);
    },
    onError: (error) => toast.error("Figma", { description: error.message }),
  });
};

export const useRefreshFigmaStoryLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  return useMutation({
    mutationFn: async ({
      storyId,
      linkId,
    }: {
      storyId: string;
      linkId: string;
    }) => {
      const response = await refreshFigmaStoryLinkAction(
        workspaceSlug,
        storyId,
        linkId,
      );
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message ??
            "The Figma preview could not be refreshed.",
        );
      }
      return response.data;
    },
    onSuccess: async (_, { storyId }) => {
      await queryClient.invalidateQueries({
        queryKey: figmaKeys.storyLinks(workspaceSlug, storyId),
      });
      toast.success("Figma preview is up to date");
    },
    onError: (error) => toast.error("Figma", { description: error.message }),
  });
};
