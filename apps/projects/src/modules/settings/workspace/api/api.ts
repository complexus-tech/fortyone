import { get, post, put, remove } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import type {
  CreateCredentialInput,
  CreatedWebhookEndpoint,
  Credential,
  DeveloperScope,
  IssuedCredential,
  IssuedOAuthApplication,
  IssuedOAuthSecret,
  OAuthApplication,
  OAuthClientSecret,
  RotatedWebhookSecret,
  ServiceAccount,
  WebhookEndpointPage,
  WebhookEventType,
} from "./types";

export const listPersonalTokens = async (ctx: WorkspaceCtx) =>
  (await get<ApiResponse<Credential[]>>("personal-access-tokens", ctx)).data ??
  [];

export const createPersonalToken = async (
  ctx: WorkspaceCtx,
  input: CreateCredentialInput,
) =>
  (
    await post<CreateCredentialInput, ApiResponse<IssuedCredential>>(
      "personal-access-tokens",
      input,
      ctx,
    )
  ).data!;

export const rotatePersonalToken = async (
  ctx: WorkspaceCtx,
  credentialId: string,
  expiresAt: string,
) =>
  (
    await post<{ expiresAt: string }, ApiResponse<IssuedCredential>>(
      `personal-access-tokens/${credentialId}/rotate`,
      { expiresAt },
      ctx,
    )
  ).data!;

export const revokePersonalToken = (ctx: WorkspaceCtx, credentialId: string) =>
  remove<ApiResponse<null>>(`personal-access-tokens/${credentialId}`, ctx);

export const listServiceAccounts = async (ctx: WorkspaceCtx) =>
  (await get<ApiResponse<ServiceAccount[]>>("service-accounts", ctx)).data ??
  [];

export const createServiceAccount = async (
  ctx: WorkspaceCtx,
  input: { name: string; workspaceRole: "guest" | "member" },
) =>
  (
    await post<typeof input, ApiResponse<ServiceAccount>>(
      "service-accounts",
      input,
      ctx,
    )
  ).data!;

export const disableServiceAccount = (ctx: WorkspaceCtx, accountId: string) =>
  remove<ApiResponse<null>>(`service-accounts/${accountId}`, ctx);

export const listServiceAccountKeys = async (
  ctx: WorkspaceCtx,
  accountId: string,
) =>
  (
    await get<ApiResponse<Credential[]>>(
      `service-accounts/${accountId}/keys`,
      ctx,
    )
  ).data ?? [];

export const createServiceAccountKey = async (
  ctx: WorkspaceCtx,
  accountId: string,
  input: CreateCredentialInput,
) =>
  (
    await post<CreateCredentialInput, ApiResponse<IssuedCredential>>(
      `service-accounts/${accountId}/keys`,
      input,
      ctx,
    )
  ).data!;

export const rotateServiceAccountKey = async (
  ctx: WorkspaceCtx,
  accountId: string,
  credentialId: string,
  expiresAt: string,
) =>
  (
    await post<
      { expiresAt: string; overlapSeconds: number },
      ApiResponse<IssuedCredential>
    >(
      `service-accounts/${accountId}/keys/${credentialId}/rotate`,
      { expiresAt, overlapSeconds: 3600 },
      ctx,
    )
  ).data!;

export const revokeServiceAccountKey = (
  ctx: WorkspaceCtx,
  accountId: string,
  credentialId: string,
) =>
  remove<ApiResponse<null>>(
    `service-accounts/${accountId}/keys/${credentialId}`,
    ctx,
  );

export const listOAuthApplications = async (ctx: WorkspaceCtx) =>
  (await get<ApiResponse<OAuthApplication[]>>("oauth-applications", ctx))
    .data ?? [];

export const createOAuthApplication = async (
  ctx: WorkspaceCtx,
  input: {
    name: string;
    redirectUris: string[];
    expiresAt: string;
    secretExpiresAt: string;
  },
) =>
  (
    await post<typeof input, ApiResponse<IssuedOAuthApplication>>(
      "oauth-applications",
      input,
      ctx,
    )
  ).data!;

export const listOAuthClientSecrets = async (
  ctx: WorkspaceCtx,
  applicationId: string,
) =>
  (
    await get<ApiResponse<OAuthClientSecret[]>>(
      `oauth-applications/${applicationId}/secrets`,
      ctx,
    )
  ).data ?? [];

export const rotateOAuthClientSecret = async (
  ctx: WorkspaceCtx,
  applicationId: string,
  expiresAt: string,
) =>
  (
    await post<
      { expiresAt: string; overlapSeconds: number },
      ApiResponse<IssuedOAuthSecret>
    >(
      `oauth-applications/${applicationId}/secrets/rotate`,
      { expiresAt, overlapSeconds: 3600 },
      ctx,
    )
  ).data!;

export const revokeOAuthClientSecret = (
  ctx: WorkspaceCtx,
  applicationId: string,
  secretId: string,
) =>
  remove<ApiResponse<null>>(
    `oauth-applications/${applicationId}/secrets/${secretId}`,
    ctx,
  );

export const listWebhookEndpoints = async (
  ctx: WorkspaceCtx,
  cursor?: string,
) => {
  const query = new URLSearchParams({ limit: "25" });
  if (cursor) query.set("cursor", cursor);
  return (
    await get<ApiResponse<WebhookEndpointPage>>(
      `webhook-endpoints?${query.toString()}`,
      ctx,
    )
  ).data!;
};

export const createWebhookEndpoint = async (
  ctx: WorkspaceCtx,
  input: { name: string; url: string; subscriptions: WebhookEventType[] },
) =>
  (
    await post<typeof input, ApiResponse<CreatedWebhookEndpoint>>(
      "webhook-endpoints",
      input,
      ctx,
    )
  ).data!;

export const replaceWebhookSubscriptions = (
  ctx: WorkspaceCtx,
  endpointId: string,
  subscriptions: WebhookEventType[],
) =>
  put<{ subscriptions: WebhookEventType[] }, ApiResponse<null>>(
    `webhook-endpoints/${endpointId}/subscriptions`,
    { subscriptions },
    ctx,
  );

export const rotateWebhookSecret = async (
  ctx: WorkspaceCtx,
  endpointId: string,
) =>
  (
    await post<Record<string, never>, ApiResponse<RotatedWebhookSecret>>(
      `webhook-endpoints/${endpointId}/rotate-secret`,
      {},
      ctx,
    )
  ).data!;

export const disableWebhookEndpoint = (ctx: WorkspaceCtx, endpointId: string) =>
  post<{ reason: string }, ApiResponse<null>>(
    `webhook-endpoints/${endpointId}/disable`,
    { reason: "disabled_from_settings" },
    ctx,
  );

export type { DeveloperScope };
