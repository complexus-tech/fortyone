"use client";

import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import {
  createOAuthApplication,
  createPersonalToken,
  createServiceAccount,
  createServiceAccountKey,
  createWebhookEndpoint,
  disableServiceAccount,
  disableWebhookEndpoint,
  listOAuthApplications,
  listOAuthClientSecrets,
  listPersonalTokens,
  listServiceAccountKeys,
  listServiceAccounts,
  listWebhookEndpoints,
  replaceWebhookSubscriptions,
  revokeOAuthClientSecret,
  revokePersonalToken,
  revokeServiceAccountKey,
  rotateOAuthClientSecret,
  rotatePersonalToken,
  rotateServiceAccountKey,
  rotateWebhookSecret,
} from "./api";
import type { CreateCredentialInput, WebhookEventType } from "./types";

const developerKeys = {
  root: (workspaceSlug: string) =>
    ["developer-settings", workspaceSlug] as const,
  personalTokens: (workspaceSlug: string) =>
    [...developerKeys.root(workspaceSlug), "personal-tokens"] as const,
  serviceAccounts: (workspaceSlug: string) =>
    [...developerKeys.root(workspaceSlug), "service-accounts"] as const,
  serviceAccountKeys: (workspaceSlug: string, accountId: string) =>
    [
      ...developerKeys.serviceAccounts(workspaceSlug),
      accountId,
      "keys",
    ] as const,
  oauthApplications: (workspaceSlug: string) =>
    [...developerKeys.root(workspaceSlug), "oauth-applications"] as const,
  oauthSecrets: (workspaceSlug: string, applicationId: string) =>
    [
      ...developerKeys.oauthApplications(workspaceSlug),
      applicationId,
      "secrets",
    ] as const,
  webhooks: (workspaceSlug: string) =>
    [...developerKeys.root(workspaceSlug), "webhooks"] as const,
};

const useDeveloperContext = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return {
    ctx: { session: session!, workspaceSlug },
    enabled: Boolean(session && workspaceSlug),
    workspaceSlug,
  };
};

export const usePersonalTokens = () => {
  const { ctx, enabled, workspaceSlug } = useDeveloperContext();
  return useQuery({
    queryKey: developerKeys.personalTokens(workspaceSlug),
    queryFn: () => listPersonalTokens(ctx),
    enabled,
  });
};

export const usePersonalTokenMutations = () => {
  const queryClient = useQueryClient();
  const { ctx, workspaceSlug } = useDeveloperContext();
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: developerKeys.personalTokens(workspaceSlug),
    });
  return {
    create: useMutation({
      mutationFn: (input: CreateCredentialInput) =>
        createPersonalToken(ctx, input),
      onSuccess: invalidate,
    }),
    rotate: useMutation({
      mutationFn: ({
        credentialId,
        expiresAt,
      }: {
        credentialId: string;
        expiresAt: string;
      }) => rotatePersonalToken(ctx, credentialId, expiresAt),
      onSuccess: invalidate,
    }),
    revoke: useMutation({
      mutationFn: (credentialId: string) =>
        revokePersonalToken(ctx, credentialId),
      onSuccess: invalidate,
    }),
  };
};

export const useServiceAccounts = () => {
  const { ctx, enabled, workspaceSlug } = useDeveloperContext();
  return useQuery({
    queryKey: developerKeys.serviceAccounts(workspaceSlug),
    queryFn: () => listServiceAccounts(ctx),
    enabled,
  });
};

export const useServiceAccountMutations = () => {
  const queryClient = useQueryClient();
  const { ctx, workspaceSlug } = useDeveloperContext();
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: developerKeys.serviceAccounts(workspaceSlug),
    });
  return {
    create: useMutation({
      mutationFn: (input: {
        name: string;
        workspaceRole: "guest" | "member";
      }) => createServiceAccount(ctx, input),
      onSuccess: invalidate,
    }),
    disable: useMutation({
      mutationFn: (accountId: string) => disableServiceAccount(ctx, accountId),
      onSuccess: invalidate,
    }),
  };
};

export const useServiceAccountKeys = (accountId: string, enabled: boolean) => {
  const { ctx, workspaceSlug } = useDeveloperContext();
  return useQuery({
    queryKey: developerKeys.serviceAccountKeys(workspaceSlug, accountId),
    queryFn: () => listServiceAccountKeys(ctx, accountId),
    enabled: enabled && Boolean(accountId),
  });
};

export const useServiceAccountKeyMutations = (accountId: string) => {
  const queryClient = useQueryClient();
  const { ctx, workspaceSlug } = useDeveloperContext();
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: developerKeys.serviceAccountKeys(workspaceSlug, accountId),
    });
  return {
    create: useMutation({
      mutationFn: (input: CreateCredentialInput) =>
        createServiceAccountKey(ctx, accountId, input),
      onSuccess: invalidate,
    }),
    rotate: useMutation({
      mutationFn: ({
        credentialId,
        expiresAt,
      }: {
        credentialId: string;
        expiresAt: string;
      }) => rotateServiceAccountKey(ctx, accountId, credentialId, expiresAt),
      onSuccess: invalidate,
    }),
    revoke: useMutation({
      mutationFn: (credentialId: string) =>
        revokeServiceAccountKey(ctx, accountId, credentialId),
      onSuccess: invalidate,
    }),
  };
};

export const useOAuthApplications = () => {
  const { ctx, enabled, workspaceSlug } = useDeveloperContext();
  return useQuery({
    queryKey: developerKeys.oauthApplications(workspaceSlug),
    queryFn: () => listOAuthApplications(ctx),
    enabled,
  });
};

export const useOAuthApplicationMutations = () => {
  const queryClient = useQueryClient();
  const { ctx, workspaceSlug } = useDeveloperContext();
  return {
    create: useMutation({
      mutationFn: (input: {
        name: string;
        redirectUris: string[];
        expiresAt: string;
        secretExpiresAt: string;
      }) => createOAuthApplication(ctx, input),
      onSuccess: () =>
        queryClient.invalidateQueries({
          queryKey: developerKeys.oauthApplications(workspaceSlug),
        }),
    }),
  };
};

export const useOAuthClientSecrets = (
  applicationId: string,
  enabled: boolean,
) => {
  const { ctx, workspaceSlug } = useDeveloperContext();
  return useQuery({
    queryKey: developerKeys.oauthSecrets(workspaceSlug, applicationId),
    queryFn: () => listOAuthClientSecrets(ctx, applicationId),
    enabled: enabled && Boolean(applicationId),
  });
};

export const useOAuthClientSecretMutations = (applicationId: string) => {
  const queryClient = useQueryClient();
  const { ctx, workspaceSlug } = useDeveloperContext();
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: developerKeys.oauthSecrets(workspaceSlug, applicationId),
    });
  return {
    rotate: useMutation({
      mutationFn: (expiresAt: string) =>
        rotateOAuthClientSecret(ctx, applicationId, expiresAt),
      onSuccess: invalidate,
    }),
    revoke: useMutation({
      mutationFn: (secretId: string) =>
        revokeOAuthClientSecret(ctx, applicationId, secretId),
      onSuccess: invalidate,
    }),
  };
};

export const useWebhookEndpoints = () => {
  const { ctx, enabled, workspaceSlug } = useDeveloperContext();
  return useInfiniteQuery({
    queryKey: developerKeys.webhooks(workspaceSlug),
    queryFn: ({ pageParam }) => listWebhookEndpoints(ctx, pageParam),
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    initialPageParam: undefined as string | undefined,
    enabled,
  });
};

export const useWebhookMutations = () => {
  const queryClient = useQueryClient();
  const { ctx, workspaceSlug } = useDeveloperContext();
  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: developerKeys.webhooks(workspaceSlug),
    });
  return {
    create: useMutation({
      mutationFn: (input: {
        name: string;
        url: string;
        subscriptions: WebhookEventType[];
      }) => createWebhookEndpoint(ctx, input),
      onSuccess: invalidate,
    }),
    replaceSubscriptions: useMutation({
      mutationFn: ({
        endpointId,
        subscriptions,
      }: {
        endpointId: string;
        subscriptions: WebhookEventType[];
      }) => replaceWebhookSubscriptions(ctx, endpointId, subscriptions),
      onSuccess: invalidate,
    }),
    rotate: useMutation({
      mutationFn: (endpointId: string) => rotateWebhookSecret(ctx, endpointId),
      onSuccess: invalidate,
    }),
    disable: useMutation({
      mutationFn: (endpointId: string) =>
        disableWebhookEndpoint(ctx, endpointId),
      onSuccess: invalidate,
    }),
  };
};
