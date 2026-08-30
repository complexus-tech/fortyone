"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Box, Container, Divider, Text, TextEditor } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { NotFoundIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { BodyContainer } from "@/components/shared/body";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useMembers } from "@/lib/hooks/members";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { RequestActivity } from "./components/request-activity";
import { RequestIntegrationBanner } from "./components/request-integration-banner";
import { RequestMetadata } from "./components/request-metadata";
import { RequestProperties } from "./components/request-properties";
import { getRequestSourceBanner } from "./components/request-source-banner";
import { useAcceptIntegrationRequest } from "./hooks/use-accept-request";
import { useDeclineIntegrationRequest } from "./hooks/use-decline-request";
import { useRequestEditors } from "./hooks/use-request-editors";
import { useIntegrationRequest } from "./hooks/use-request";
import { useUpdateIntegrationRequest } from "./hooks/use-update-request";
import type { UpdateIntegrationRequestInput } from "./types";
import { getAcceptedStoryPath } from "./utils/accepted-story-path";

export { RequestIntegrationBanner };

const metadataText = (value: unknown) =>
  typeof value === "string" && value.trim() ? value : null;

export const IntegrationRequestDetails = ({
  requestId,
}: {
  requestId: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const [isDeclining, setIsDeclining] = useState(false);
  const { data: request, isPending } = useIntegrationRequest(requestId);
  const requestTeamId = request?.teamId ?? teamId;
  const { data: statuses = [] } = useTeamStatuses(requestTeamId);
  const { data: members = [] } = useMembers();
  const { mutate: updateRequest } = useUpdateIntegrationRequest();
  const acceptRequest = useAcceptIntegrationRequest();
  const declineRequest = useDeclineIntegrationRequest();

  const defaultStatus =
    statuses.find((status) => status.category === "unstarted") ||
    statuses.at(0);
  const statusId = request?.statusId ?? defaultStatus?.id;
  const priority = request?.priority ?? "No Priority";

  const handleUpdate = (payload: UpdateIntegrationRequestInput) => {
    if (!request?.id) return;
    updateRequest({ requestId: request.id, payload });
  };
  const { descriptionEditor, titleEditor } = useRequestEditors({
    onUpdate: handleUpdate,
    request,
    requestId,
  });

  if (isPending) {
    return (
      <Box className="h-full px-8 py-7">
        <Box className="bg-surface-muted mb-8 h-8 w-2/5 rounded" />
        <Box className="bg-surface-muted mb-4 h-4 w-4/5 rounded" />
        <Box className="bg-surface-muted h-4 w-3/5 rounded" />
      </Box>
    );
  }

  if (!request) {
    return (
      <Box className="flex h-full items-center justify-center px-6">
        <Box className="flex flex-col items-center">
          <NotFoundIllustration className="mb-5 w-52" />
          <Text align="center" className="mb-3" fontSize="xl">
            Intake item not found
          </Text>
          <Text align="center" color="muted">
            This intake item may have already been handled.
          </Text>
        </Box>
      </Box>
    );
  }

  const repositoryName = metadataText(request.metadata.repository_full_name);
  const slackChannel = metadataText(request.metadata.slack_channel);
  const issueNumber = request.sourceNumber ? `#${request.sourceNumber}` : "";
  const sourceBanner = getRequestSourceBanner({
    issueNumber,
    provider: request.provider,
    repositoryName,
    slackChannel,
    storyTerm: getTermDisplay("storyTerm", { capitalize: true }),
  });
  const selectedStatus = statuses.find((status) => status.id === statusId);
  const assignee = members.find((member) => member.id === request.assigneeId);
  const canEditRequest = request.status === "pending";

  const handleAccept = () => {
    acceptRequest.mutate(request.id, {
      onSuccess: (res) => {
        if (res.data?.acceptedStoryId) {
          router.push(
            withWorkspace(getAcceptedStoryPath(res.data.acceptedStoryId)),
          );
        }
      },
    });
  };

  return (
    <Box className="h-full min-h-0">
      <Box className="notification-story-container flex h-full min-h-0 flex-col overflow-hidden md:flex-row">
        <Box className="min-h-0 min-w-0 flex-1">
          <BodyContainer className="h-full min-h-0 overflow-y-auto pb-8">
            <Container className="max-w-7xl pt-7">
              <RequestIntegrationBanner
                canEditRequest={canEditRequest}
                icon={sourceBanner.icon}
                onAccept={handleAccept}
                onDecline={() => {
                  setIsDeclining(true);
                }}
                openLabel={sourceBanner.openLabel}
                primaryText={sourceBanner.primaryText}
                secondaryText={sourceBanner.secondaryText}
                sourceUrl={request.sourceUrl}
              />
              <TextEditor
                asTitle
                className="text-foreground mb-8 text-3xl md:text-4xl"
                editor={titleEditor}
              />
              <TextEditor className="text-lg" editor={descriptionEditor} />
              <Box className="notification-story-inline-options mt-6 hidden">
                <RequestProperties
                  assignee={assignee}
                  canEditRequest={canEditRequest}
                  onUpdate={handleUpdate}
                  priority={priority}
                  request={request}
                  selectedStatus={selectedStatus}
                  statusId={statusId}
                  teamId={request.teamId}
                  variant="inline"
                />
              </Box>
              <RequestMetadata metadata={request.metadata} />
              <Divider className="my-6" />
              <RequestActivity
                provider={request.provider}
                requestId={request.id}
              />
            </Container>
          </BodyContainer>
        </Box>

        <Box className="notification-story-sidebar from-sidebar/70 to-sidebar/40 border-border w-full shrink-0 overflow-y-auto border-t-[0.5px] bg-linear-to-br pb-6 md:h-full md:min-h-0 md:w-(--story-sidebar-width) md:border-t-0 md:border-l-[0.5px]">
          <RequestProperties
            assignee={assignee}
            canEditRequest={canEditRequest}
            onUpdate={handleUpdate}
            priority={priority}
            request={request}
            selectedStatus={selectedStatus}
            statusId={statusId}
            teamId={request.teamId}
          />
        </Box>
      </Box>
      <ConfirmDialog
        confirmText="Decline intake item"
        description="Declining removes this item from the team's intake queue. You can still find the original item in the source integration."
        isLoading={declineRequest.isPending}
        isOpen={isDeclining}
        loadingText="Declining..."
        onCancel={() => {
          setIsDeclining(false);
        }}
        onClose={() => {
          setIsDeclining(false);
        }}
        onConfirm={() => {
          declineRequest.mutate(request.id, {
            onSuccess: (res) => {
              if (!res.error?.message) {
                setIsDeclining(false);
                router.push(withWorkspace(`/teams/${request.teamId}/requests`));
              }
            },
          });
        }}
        title="Decline this intake item?"
      />
    </Box>
  );
};
