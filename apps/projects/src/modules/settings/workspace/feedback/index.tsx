"use client";

import { useState } from "react";
import { cn } from "lib";
import {
  ArrowDownIcon,
  CodeIcon,
  DeleteIcon,
  ExternalLinkIcon,
  KanbanIcon,
  MoreHorizontalIcon,
  PlusIcon,
  SettingsIcon,
  TeamIcon,
  UpdatesIcon,
} from "icons";
import {
  Box,
  Button,
  Dialog,
  Flex,
  Input,
  Menu,
  Select,
  Switch,
  Tabs,
  Text,
} from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import type {
  FeedbackBoard,
  FeedbackGuestIdentityPolicy,
  FeedbackParticipationMode,
  FeedbackPortal,
} from "./types";
import {
  useCreateFeedbackBoardMutation,
  useDeleteFeedbackBoardMutation,
  useFeedbackPortals,
  useUpdateFeedbackPortalMutation,
} from "./hooks";
import { FeedbackReviewersDialog } from "./reviewers-dialog";
import { WidgetInstallSettings } from "./widget-install-settings";
import { FeedbackUpdatesSettings } from "./updates-settings";

const colorOptions = [
  { label: "Green", value: "green" },
  { label: "Blue", value: "blue" },
  { label: "Yellow", value: "yellow" },
  { label: "Pink", value: "pink" },
  { label: "Red", value: "red" },
];

const participationOptions: {
  description: string;
  label: string;
  value: FeedbackParticipationMode;
}[] = [
  {
    description:
      "People must log in or create a FortyOne account before submitting feedback.",
    label: "Account required",
    value: "account_required",
  },
  {
    description:
      "People can verify their email without creating a FortyOne account, then receive replies and progress updates.",
    label: "Verified email",
    value: "verified_guest",
  },
  {
    description:
      "People can choose verified email for updates or submit without a name, email address, or personal notifications.",
    label: "Anonymous allowed",
    value: "anonymous_allowed",
  },
];

const guestIdentityOptions: {
  description: string;
  label: string;
  value: FeedbackGuestIdentityPolicy;
}[] = [
  {
    description:
      "Show a verified guest’s display name publicly. Their email address remains private.",
    label: "Show display name",
    value: "show_identity",
  },
  {
    description:
      "Let each verified guest decide whether their display name is public or shown as Anonymous.",
    label: "Let guests choose",
    value: "allow_public_masking",
  },
  {
    description:
      "Show every verified guest as Anonymous publicly while administrators retain their verified identity.",
    label: "Always hide names",
    value: "always_mask_guests",
  },
];

const isWorkspaceSubdomainDeployment =
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

const selectClassName =
  "border-border ring-ring h-11 w-full appearance-none rounded-xl border bg-white pr-10 pl-3 text-sm outline-none focus-visible:ring-2 dark:bg-surface-elevated";

const PublicUrl = ({ portal }: { portal: FeedbackPortal }) => {
  return (
    <Button
      color="tertiary"
      href={
        isWorkspaceSubdomainDeployment
          ? "/feedback"
          : `/portal/${portal.slug}/feedback`
      }
      rel="noreferrer"
      rightIcon={<ExternalLinkIcon className="h-4" />}
      size="sm"
      target="_blank"
      variant="naked"
    >
      Open public portal
    </Button>
  );
};

const PortalConfiguration = ({
  anonymousFeedbackAvailable,
  portal,
}: {
  anonymousFeedbackAvailable: boolean;
  portal: FeedbackPortal;
}) => {
  const [isPublic, setIsPublic] = useState(portal.isPublic);
  const [participationMode, setParticipationMode] = useState(
    portal.participationMode,
  );
  const [guestIdentityPolicy, setGuestIdentityPolicy] = useState(
    portal.guestIdentityPolicy,
  );
  const mutation = useUpdateFeedbackPortalMutation();
  const selectedParticipation =
    participationOptions.find((option) => option.value === participationMode) ??
    participationOptions[0];
  const selectedGuestIdentity =
    guestIdentityOptions.find(
      (option) => option.value === guestIdentityPolicy,
    ) ?? guestIdentityOptions[0];

  const updateAvailability = async (checked: boolean) => {
    const previousValue = isPublic;
    setIsPublic(checked);

    try {
      const response = await mutation.mutateAsync({
        portalId: portal.id,
        input: {
          isPublic: checked,
        },
      });
      if (response.error?.message) {
        setIsPublic(previousValue);
      }
    } catch {
      setIsPublic(previousValue);
    }
  };

  const updateParticipationMode = async (
    nextMode: FeedbackParticipationMode,
  ) => {
    if (nextMode === "anonymous_allowed" && !anonymousFeedbackAvailable) {
      return;
    }
    const previousMode = participationMode;
    setParticipationMode(nextMode);

    try {
      const response = await mutation.mutateAsync({
        portalId: portal.id,
        input: {
          participationMode: nextMode,
        },
      });
      if (response.error?.message) {
        setParticipationMode(previousMode);
      }
    } catch {
      setParticipationMode(previousMode);
    }
  };

  const updateGuestIdentityPolicy = async (
    nextPolicy: FeedbackGuestIdentityPolicy,
  ) => {
    const previousPolicy = guestIdentityPolicy;
    setGuestIdentityPolicy(nextPolicy);

    try {
      const response = await mutation.mutateAsync({
        portalId: portal.id,
        input: { guestIdentityPolicy: nextPolicy },
      });
      if (response.error?.message) {
        setGuestIdentityPolicy(previousPolicy);
      }
    } catch {
      setGuestIdentityPolicy(previousPolicy);
    }
  };

  return (
    <Box className="border-border bg-surface mb-6 rounded-2xl border">
      <SectionHeader
        description="Control access to your public feedback portal and choose who can participate."
        title="General Information"
      />
      <Box className="divide-border divide-y-[0.5px]">
        <Flex align="center" className="gap-3 px-6 py-4" justify="between">
          <Box>
            <Text>Enabled</Text>
            <Text color="muted" fontSize="sm">
              When enabled, people can view and submit feedback on the public
              portal.
            </Text>
          </Box>
          <Switch
            aria-label="Enable public feedback portal"
            checked={isPublic}
            disabled={mutation.isPending}
            onCheckedChange={(checked) => {
              void updateAvailability(checked);
            }}
          />
        </Flex>
        <Flex align="center" className="gap-4 px-6 py-4" justify="between">
          <Box className="min-w-0 flex-1">
            <Text>Who can submit feedback</Text>
            <Text className="max-w-xl" color="muted" fontSize="sm">
              {selectedParticipation.description}
            </Text>
          </Box>
          <Select
            disabled={mutation.isPending}
            onValueChange={(value) => {
              void updateParticipationMode(value as FeedbackParticipationMode);
            }}
            value={participationMode}
          >
            <Select.Trigger
              aria-label="Who can submit feedback"
              className="w-max shrink-0 text-[0.9rem] md:text-base"
            >
              <Select.Input />
            </Select.Trigger>
            <Select.Content align="end">
              {participationOptions.map((option) => (
                <Select.Option
                  className="text-base"
                  disabled={
                    option.value === "anonymous_allowed" &&
                    !anonymousFeedbackAvailable
                  }
                  key={option.value}
                  value={option.value}
                >
                  {option.label}
                </Select.Option>
              ))}
            </Select.Content>
          </Select>
        </Flex>
        {participationMode !== "account_required" ? (
          <Flex align="center" className="gap-4 px-6 py-4" justify="between">
            <Box className="min-w-0 flex-1">
              <Text>Guest public identity</Text>
              <Text className="max-w-xl" color="muted" fontSize="sm">
                {selectedGuestIdentity.description}
              </Text>
            </Box>
            <Select
              disabled={mutation.isPending}
              onValueChange={(value) => {
                void updateGuestIdentityPolicy(
                  value as FeedbackGuestIdentityPolicy,
                );
              }}
              value={guestIdentityPolicy}
            >
              <Select.Trigger
                aria-label="Guest public identity"
                className="w-max shrink-0 text-[0.9rem] md:text-base"
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content align="end">
                {guestIdentityOptions.map((option) => (
                  <Select.Option
                    className="text-base"
                    key={option.value}
                    value={option.value}
                  >
                    {option.label}
                  </Select.Option>
                ))}
              </Select.Content>
            </Select>
          </Flex>
        ) : null}
        {!anonymousFeedbackAvailable ? (
          <Box className="px-6 py-4">
            <Text color="muted" fontSize="sm">
              Anonymous feedback is unavailable on this deployment. Configure a
              trusted reverse proxy and FEEDBACK_TRUSTED_CLIENT_IP_HEADER before
              enabling it.
            </Text>
          </Box>
        ) : null}
      </Box>
    </Box>
  );
};

const CreateBoardDialog = ({ portal }: { portal?: FeedbackPortal }) => {
  const { data: teams = [] } = useTeams();
  const teamsWithBoards = new Set(
    (portal?.boards ?? []).map((board) => board.teamId),
  );
  const availableTeams = teams.filter((team) => !teamsWithBoards.has(team.id));
  const [open, setOpen] = useState(false);
  const [teamId, setTeamId] = useState(availableTeams[0]?.id ?? "");
  const [name, setName] = useState("");
  const [color, setColor] = useState("green");
  const mutation = useCreateFeedbackBoardMutation();

  const submit = async () => {
    const response = await mutation.mutateAsync({
      color,
      name,
      portalId: portal?.id ?? "",
      teamId,
    });
    if (response.error?.message) return;
    setOpen(false);
    setName("");
  };

  return (
    <Dialog onOpenChange={setOpen} open={open}>
      <Button
        color="tertiary"
        disabled={!portal || availableTeams.length === 0}
        leftIcon={<PlusIcon className="h-[1.1rem]" />}
        onClick={() => {
          setTeamId(availableTeams[0]?.id ?? "");
          setOpen(true);
        }}
        size="sm"
      >
        Create Board
      </Button>
      <Dialog.Content className="max-w-lg">
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-1 text-lg">
            Create Feedback Board
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Box>
            <Text className="mb-2 text-sm" fontWeight="medium">
              Owning team
            </Text>
            <Box className="relative">
              <select
                aria-label="Owning team"
                className={selectClassName}
                onChange={(event) => {
                  setTeamId(event.target.value);
                }}
                value={teamId}
              >
                {availableTeams.map((team) => (
                  <option key={team.id} value={team.id}>
                    {team.name}
                  </option>
                ))}
              </select>
              <ArrowDownIcon className="text-text-muted pointer-events-none absolute top-1/2 right-3.5 h-3.5 w-auto -translate-y-1/2" />
            </Box>
          </Box>
          <Input
            onChange={(event) => {
              setName(event.target.value);
            }}
            placeholder="Board name"
            value={name}
          />
          <Flex className="flex-wrap gap-2">
            {colorOptions.map((option) => (
              <button
                aria-pressed={color === option.value}
                className={cn(
                  "bg-surface-muted hover:bg-state-hover flex items-center gap-2 rounded-lg border px-3 py-2 text-sm transition",
                  color === option.value
                    ? "border-primary/40 bg-primary/5"
                    : "border-transparent",
                )}
                key={option.value}
                onClick={() => {
                  setColor(option.value);
                }}
                type="button"
              >
                <Box
                  aria-hidden="true"
                  className="size-3 rounded-sm"
                  style={{ backgroundColor: option.value }}
                />
                {option.label}
              </button>
            ))}
          </Flex>
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3">
          <Button
            color="tertiary"
            onClick={() => {
              setOpen(false);
            }}
          >
            Cancel
          </Button>
          <Button
            color="primary"
            disabled={
              !portal ||
              !teamId ||
              name.trim().length === 0 ||
              mutation.isPending
            }
            onClick={() => {
              void submit();
            }}
          >
            Create
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export const FeedbackSettings = ({
  anonymousFeedbackAvailable = true,
}: {
  anonymousFeedbackAvailable?: boolean;
}) => {
  const { data: portals = [], isLoading } = useFeedbackPortals();
  const { data: teams = [] } = useTeams();
  const [reviewersDialog, setReviewersDialog] = useState<{
    board: FeedbackBoard;
    teamName: string;
  } | null>(null);
  const [boardToDelete, setBoardToDelete] = useState<FeedbackBoard | null>(
    null,
  );
  const deleteBoard = useDeleteFeedbackBoardMutation();
  const primaryPortal = portals.at(0);

  const teamsById = new Map(teams.map((team) => [team.id, team]));

  const boards = portals.flatMap((portal) =>
    (portal.boards ?? []).map((board) => ({
      ...board,
      portalName: portal.name,
      team: teamsById.get(board.teamId),
    })),
  );

  return (
    <Box>
      <Flex align="center" className="mb-5" gap={3} justify="between">
        <Text as="h1" className="text-2xl font-medium">
          Feedback
        </Text>
        {primaryPortal ? <PublicUrl portal={primaryPortal} /> : null}
      </Flex>

      <Text className="mb-6 max-w-2xl leading-relaxed" color="muted">
        Collect requests, votes, and product feedback in one public portal.
        Boards route each submission to the right team.
      </Text>

      <Tabs defaultValue="general">
        <Box className="overflow-x-auto">
          <Tabs.List className="mx-0 md:mx-0">
            <Tabs.Tab
              leftIcon={<SettingsIcon className="h-[1.1rem]" />}
              value="general"
            >
              General
            </Tabs.Tab>
            <Tabs.Tab
              leftIcon={<KanbanIcon className="h-[1.1rem]" />}
              value="boards"
            >
              Boards
            </Tabs.Tab>
            <Tabs.Tab
              leftIcon={<CodeIcon className="h-[1.1rem]" />}
              value="widget"
            >
              Widget
            </Tabs.Tab>
            <Tabs.Tab
              leftIcon={<UpdatesIcon className="h-[1.1rem]" />}
              value="updates"
            >
              Updates
            </Tabs.Tab>
          </Tabs.List>
        </Box>

        <Box className="mt-5">
          <Tabs.Panel value="general">
            {primaryPortal ? (
              <PortalConfiguration
                anonymousFeedbackAvailable={anonymousFeedbackAvailable}
                key={`${primaryPortal.id}:${primaryPortal.isPublic}:${primaryPortal.participationMode}:${primaryPortal.guestIdentityPolicy}`}
                portal={primaryPortal}
              />
            ) : null}
          </Tabs.Panel>
          <Tabs.Panel value="boards">
            <Box className="border-border bg-surface rounded-2xl border">
              <SectionHeader
                action={<CreateBoardDialog portal={primaryPortal} />}
                description="Boards route public feedback to the team that owns the work."
                title="Boards"
              />
              {boards.length === 0 && !isLoading ? (
                <Flex
                  align="center"
                  className="px-6 py-10"
                  direction="column"
                  justify="center"
                >
                  <KanbanIcon className="h-12 w-auto" />
                  <Text className="mt-4 text-lg font-semibold">
                    No boards found
                  </Text>
                  <Text className="mb-3" color="muted">
                    Create a team-linked board to start collecting feedback.
                  </Text>
                </Flex>
              ) : (
                <>
                  <Flex
                    align="center"
                    className="border-border hidden border-b-[0.5px] px-6 py-5 md:flex"
                    justify="between"
                  >
                    <Text>Board</Text>
                    <Flex align="center" gap={3} justify="between">
                      <Text className="w-40">Team</Text>
                      <Text className="w-40">Portal</Text>
                      <Text className="w-40 text-right">Actions</Text>
                    </Flex>
                  </Flex>
                  {boards.map((board) => (
                    <Flex
                      align="center"
                      className="border-border border-b px-6 py-4 last:border-b-0"
                      justify="between"
                      key={board.id}
                    >
                      <Flex align="center" gap={2}>
                        <Box
                          aria-hidden="true"
                          className="size-3 shrink-0 rounded-sm"
                          style={{ backgroundColor: board.color }}
                        />
                        <Text className="font-medium">{board.name}</Text>
                      </Flex>
                      <Flex align="center" gap={3} justify="between">
                        <Text className="w-40">
                          {board.team?.name ?? "Missing team"}
                        </Text>
                        <Text className="w-40" color="muted">
                          {board.portalName}
                        </Text>
                        <Flex
                          align="center"
                          className="w-40"
                          gap={1}
                          justify="end"
                        >
                          <Menu>
                            <Menu.Button>
                              <Button
                                aria-label={`Options for ${board.name}`}
                                asIcon
                                color="tertiary"
                                size="sm"
                                variant="naked"
                              >
                                <MoreHorizontalIcon className="h-5 w-auto" />
                              </Button>
                            </Menu.Button>
                            <Menu.Items align="end" className="w-56">
                              <Menu.Group>
                                <Menu.Item
                                  onSelect={() => {
                                    openDialogAfterMenuClose((open) => {
                                      if (!open) return;
                                      setReviewersDialog({
                                        board,
                                        teamName: board.team?.name ?? "team",
                                      });
                                    });
                                  }}
                                >
                                  <TeamIcon />
                                  Manage Reviewers
                                </Menu.Item>
                                <Menu.Item
                                  onSelect={() => {
                                    openDialogAfterMenuClose((open) => {
                                      if (open) setBoardToDelete(board);
                                    });
                                  }}
                                >
                                  <DeleteIcon />
                                  Delete Board
                                </Menu.Item>
                              </Menu.Group>
                            </Menu.Items>
                          </Menu>
                        </Flex>
                      </Flex>
                    </Flex>
                  ))}
                </>
              )}
            </Box>
          </Tabs.Panel>
          <Tabs.Panel value="widget">
            {primaryPortal ? (
              <WidgetInstallSettings portalSlug={primaryPortal.slug} />
            ) : null}
          </Tabs.Panel>
          <Tabs.Panel value="updates">
            {primaryPortal ? (
              <FeedbackUpdatesSettings portal={primaryPortal} />
            ) : null}
          </Tabs.Panel>
        </Box>
      </Tabs>
      {reviewersDialog ? (
        <FeedbackReviewersDialog
          board={reviewersDialog.board}
          onOpenChange={(open) => {
            if (!open) setReviewersDialog(null);
          }}
          open
          teamName={reviewersDialog.teamName}
        />
      ) : null}
      <ConfirmDialog
        confirmText="Delete board"
        description="This permanently deletes the board and all feedback submitted to it. This action cannot be undone."
        isLoading={deleteBoard.isPending}
        isOpen={Boolean(boardToDelete)}
        loadingText="Deleting..."
        onCancel={() => {
          setBoardToDelete(null);
        }}
        onClose={() => {
          setBoardToDelete(null);
        }}
        onConfirm={() => {
          if (!boardToDelete) return;
          deleteBoard.mutate(boardToDelete.id, {
            onSuccess: () => {
              setBoardToDelete(null);
            },
          });
        }}
        title={`Delete ${boardToDelete?.name ?? "board"}?`}
      />
    </Box>
  );
};
