import { acceptIntegrationRequestAction } from "@/modules/integration-requests/actions/accept";
import { acceptAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/accept-all";
import { declineIntegrationRequestAction } from "@/modules/integration-requests/actions/decline";
import { declineAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/decline-all";
import { postRequestGitHubCommentAction } from "@/modules/integration-requests/actions/post-github-comment";
import { updateIntegrationRequestAction } from "@/modules/integration-requests/actions/update";
import { getIntegrationRequest } from "@/modules/integration-requests/queries/get-request";
import { getRequestGitHubComments } from "@/modules/integration-requests/queries/get-request-github-comments";
import { getTeamIntegrationRequestsPage } from "@/modules/integration-requests/queries/get-team-requests";
import type { IntegrationRequestToolDependencies } from "./integration-requests/contracts";
import { createIntegrationRequestMutationTools } from "./integration-requests/mutation-tools";
import { createIntegrationRequestReadTools } from "./integration-requests/read-tools";

/**
 * The feature module owns the action/query implementations. This façade keeps
 * the lower-layer dependency explicit while grouped tools stay independent of
 * feature-internal import paths.
 */
const integrationRequestToolDependencies = {
  acceptAllIntegrationRequests: acceptAllIntegrationRequestsAction,
  acceptIntegrationRequest: acceptIntegrationRequestAction,
  declineAllIntegrationRequests: declineAllIntegrationRequestsAction,
  declineIntegrationRequest: declineIntegrationRequestAction,
  getIntegrationRequest,
  getRequestGitHubComments,
  getTeamIntegrationRequestsPage,
  postRequestGitHubComment: postRequestGitHubCommentAction,
  updateIntegrationRequest: updateIntegrationRequestAction,
} satisfies IntegrationRequestToolDependencies;

export const {
  getIntegrationRequestTool,
  getRequestGitHubCommentsTool,
  listIntegrationRequestsTool,
} = createIntegrationRequestReadTools(integrationRequestToolDependencies);

export const {
  acceptAllIntegrationRequestsTool,
  acceptIntegrationRequestTool,
  declineAllIntegrationRequestsTool,
  declineIntegrationRequestTool,
  postRequestGitHubCommentTool,
  updateIntegrationRequestTool,
} = createIntegrationRequestMutationTools(integrationRequestToolDependencies);
