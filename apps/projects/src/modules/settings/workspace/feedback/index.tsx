"use client";

import { useEffect, useMemo, useState } from "react";
import { LinkIcon, PlusIcon, TeamIcon } from "icons";
import { Box, Button, Dialog, Flex, Input, Switch, Text } from "ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import type { FeedbackPortal } from "./types";
import {
  useCreateFeedbackBoardMutation,
  useFeedbackPortals,
  useUpdateFeedbackPortalMutation,
} from "./hooks";

const colorOptions = [
  { label: "Green", value: "green" },
  { label: "Blue", value: "blue" },
  { label: "Yellow", value: "yellow" },
  { label: "Pink", value: "pink" },
  { label: "Red", value: "red" },
];

const isWorkspaceSubdomainDeployment =
  process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

const selectClassName =
  "border-border ring-ring h-11 w-full rounded-xl border bg-white px-3 text-sm outline-none focus-visible:ring-2 dark:bg-surface-elevated";

const PublicUrl = ({ portal }: { portal: FeedbackPortal }) => {
  return (
    <Button
      color="tertiary"
      href={
        isWorkspaceSubdomainDeployment
          ? "/feedback"
          : `/portal/${portal.slug}/feedback`
      }
      leftIcon={<LinkIcon className="h-4" />}
      size="sm"
      target="_blank"
    >
      Public portal
    </Button>
  );
};

const PortalConfiguration = ({ portal }: { portal: FeedbackPortal }) => {
  const [isPublic, setIsPublic] = useState(portal.isPublic);
  const mutation = useUpdateFeedbackPortalMutation();

  useEffect(() => {
    setIsPublic(portal.isPublic);
  }, [portal.isPublic]);

  const hasChanges = isPublic !== portal.isPublic;

  const submit = async () => {
    await mutation.mutateAsync({
      portalId: portal.id,
      input: {
        isPublic,
      },
    });
  };

  return (
    <Box className="border-border bg-surface mb-6 rounded-2xl border">
      <SectionHeader
        action={<PublicUrl portal={portal} />}
        description="Choose whether people can access your public feedback portal."
        title="Public portal"
      />
      <Box className="space-y-5 p-6">
        <Flex align="center" justify="between">
          <Box>
            <Text className="font-medium">Enabled</Text>
            <Text className="mt-1" color="muted">
              When enabled, people can view and submit feedback on the public
              portal.
            </Text>
          </Box>
          <Switch
            aria-label="Enable public feedback portal"
            checked={isPublic}
            onCheckedChange={(checked) => {
              setIsPublic(checked);
            }}
          />
        </Flex>
        <Flex justify="end">
          <Button
            color="primary"
            disabled={!hasChanges || mutation.isPending}
            onClick={() => {
              void submit();
            }}
          >
            Save Changes
          </Button>
        </Flex>
      </Box>
    </Box>
  );
};

const CreateBoardDialog = ({ portal }: { portal?: FeedbackPortal }) => {
  const { data: teams = [] } = useTeams();
  const [open, setOpen] = useState(false);
  const [teamId, setTeamId] = useState(teams[0]?.id ?? "");
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
        disabled={!portal || teams.length === 0}
        leftIcon={<PlusIcon className="h-[1.1rem]" />}
        onClick={() => {
          setTeamId(teams[0]?.id ?? "");
          setOpen(true);
        }}
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
            <select
              aria-label="Owning team"
              className={selectClassName}
              onChange={(event) => {
                setTeamId(event.target.value);
              }}
              value={teamId}
            >
              {teams.map((team) => (
                <option key={team.id} value={team.id}>
                  {team.name}
                </option>
              ))}
            </select>
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
                className="bg-surface-muted hover:bg-state-hover rounded-full px-3 py-2 text-sm"
                key={option.value}
                onClick={() => {
                  setColor(option.value);
                }}
                type="button"
              >
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

export const FeedbackSettings = () => {
  const { data: portals = [], isLoading } = useFeedbackPortals();
  const { data: teams = [] } = useTeams();
  const primaryPortal = portals.at(0);

  const teamsById = useMemo(
    () => new Map(teams.map((team) => [team.id, team])),
    [teams],
  );

  const boards = portals.flatMap((portal) =>
    portal.boards.map((board) => ({
      ...board,
      portalName: portal.name,
      team: teamsById.get(board.teamId),
    })),
  );

  return (
    <Box>
      <Flex align="center" className="mb-5" justify="between">
        <Text as="h1" className="text-2xl font-medium">
          Feedback
        </Text>
        <CreateBoardDialog portal={primaryPortal} />
      </Flex>

      <Text className="mb-6 max-w-2xl leading-relaxed" color="muted">
        Feedback gives your organization a public portal where customers can
        submit requests, vote on ideas, and follow progress. Boards route every
        submission to the team responsible for it.
      </Text>

      {primaryPortal ? <PortalConfiguration portal={primaryPortal} /> : null}

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
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
            <TeamIcon className="h-12 w-auto" />
            <Text className="mt-4 text-lg font-semibold">No boards found</Text>
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
              </Flex>
            </Flex>
            {boards.map((board) => (
              <Flex
                align="center"
                className="border-border border-b px-6 py-4 last:border-b-0"
                justify="between"
                key={board.id}
              >
                <Flex align="center" gap={3}>
                  <Flex
                    align="center"
                    aria-hidden="true"
                    className="bg-surface-muted/70 size-8 shrink-0 rounded-lg"
                    justify="center"
                  >
                    <Box
                      className="size-3 rounded-sm"
                      style={{ backgroundColor: board.color }}
                    />
                  </Flex>
                  <Box>
                    <Text className="font-medium">{board.name}</Text>
                    <Text className="mt-1" color="muted">
                      {board.slug}
                    </Text>
                  </Box>
                </Flex>
                <Flex align="center" gap={3} justify="between">
                  <Text className="w-40">
                    {board.team?.name ?? "Missing team"}
                  </Text>
                  <Text className="w-40" color="muted">
                    {board.portalName}
                  </Text>
                </Flex>
              </Flex>
            ))}
          </>
        )}
      </Box>
    </Box>
  );
};
