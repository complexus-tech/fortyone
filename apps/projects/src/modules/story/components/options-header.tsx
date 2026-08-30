"use client";
import { Button, Container, Flex, Text, Tooltip } from "ui";
import { CopyIcon, GitIcon, MaximizeIcon } from "icons";
import { toast } from "sonner";
import {
  useCopyToClipboard,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { useStoryById } from "@/modules/story/hooks/story";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useProfile } from "@/lib/hooks/profile";
import { getStoryPath } from "@/shared/routing/story";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useGitHubIntegration } from "@/lib/hooks/github";
import { MobileMenuButton } from "@/components/shared";
import { buildGitBranchName } from "@/modules/settings/workspace/integrations/github/branch-format";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import { StoryActionsMenu } from "./story-actions-menu";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

export const OptionsHeader = ({
  isAdminOrOwner,
  storyId,
  isDialog,
}: {
  isAdminOrOwner: boolean;
  storyId: string;
  isDialog?: boolean;
}) => {
  const { data: currentUser } = useProfile();
  const { data } = useStoryById(storyId);
  const { id, teamId, title, sequenceId, assigneeId, statusId } = data!;
  const { data: teams = [] } = useTeams();
  const [_, copyText] = useCopyToClipboard();
  const team = teams.find((team) => team.id === teamId);
  const code = team ? team.code : "";
  const { mutate: updateStory } = useUpdateStoryMutation();
  const { data: statuses } = useTeamStatuses(teamId);
  const { data: githubIntegration } = useGitHubIntegration();
  const { getTermDisplay } = useTerminology();
  const { data: automationPreferences } = useAutomationPreferences();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();

  const generateGitBranchName = () => {
    return buildGitBranchName({
      format:
        githubIntegration?.settings.branchFormat ?? "username/identifier-title",
      username: currentUser?.username,
      teamCode: code,
      sequenceId,
      title,
    });
  };

  const getStoryUrl = () => {
    const storyPath = getStoryPath({
      id,
      sequenceId,
      teamCode: code,
    });

    if (isFortyOneApp) {
      return `${window.location.origin}${storyPath}`;
    }
    return `${window.location.origin}/${workspaceSlug}${storyPath}`;
  };

  const copyBranchName = async () => {
    await copyText(generateGitBranchName());
    toast.info(generateGitBranchName(), {
      description: "Git branch name copied to clipboard",
    });

    const startedStatuses =
      statuses?.filter((status) => status.category === "started") || [];
    const updatePayload: { assigneeId?: string; statusId?: string } = {};
    const currentStatusCategory = statuses?.find(
      (status) => status.id === statusId,
    )?.category;
    if (
      automationPreferences?.assignSelfOnBranchCopy &&
      assigneeId !== currentUser?.id
    ) {
      updatePayload.assigneeId = currentUser?.id;
    }
    if (
      automationPreferences?.moveStoryToStartedOnBranch &&
      startedStatuses.length > 0 &&
      currentStatusCategory !== "started"
    ) {
      updatePayload.statusId = startedStatuses[0].id;
    }
    if (Object.keys(updatePayload).length > 0) {
      updateStory({ storyId: id, payload: updatePayload });
    }
  };

  return (
    <Container className="border-border d flex h-16 w-full items-center justify-between border-b-[0.5px] md:border-b-0 md:px-6">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <Text color="muted" fontWeight="semibold" transform="uppercase">
          {code ? (
            <>
              {code}-{sequenceId}
            </>
          ) : null}
        </Text>
      </Flex>
      <Flex align="center" gap={2}>
        {isDialog ? (
          <Tooltip side="bottom" title="Fullscreen">
            <span>
              <Button
                asIcon
                color="tertiary"
                href={withWorkspace(
                  getStoryPath({
                    id,
                    sequenceId,
                    teamCode: code,
                  }),
                )}
                leftIcon={<MaximizeIcon className="h-5" strokeWidth={2.5} />}
                variant="naked"
              >
                <span className="sr-only">Fullscreen</span>
              </Button>
            </span>
          </Tooltip>
        ) : null}

        <Tooltip
          title={`Copy ${getTermDisplay("storyTerm", { capitalize: true })} link`}
        >
          <Button
            color="tertiary"
            leftIcon={<CopyIcon />}
            onClick={async () => {
              await copyText(getStoryUrl());
              toast.info("Success", {
                description: `${getTermDisplay("storyTerm", { capitalize: true })} link copied to clipboard`,
              });
            }}
            suppressHydrationWarning
            variant="naked"
          >
            <span className="sr-only">
              Copy {getTermDisplay("storyTerm")} link
            </span>
          </Button>
        </Tooltip>
        {userRole !== "guest" && (
          <Tooltip title="Copy git branch name">
            <Button
              color="tertiary"
              disabled={!code}
              leftIcon={<GitIcon />}
              onClick={copyBranchName}
              variant="naked"
            >
              <span className="sr-only">Copy git branch name</span>
            </Button>
          </Tooltip>
        )}
        {!isDialog ? (
          <StoryActionsMenu isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
        ) : null}
      </Flex>
    </Container>
  );
};
