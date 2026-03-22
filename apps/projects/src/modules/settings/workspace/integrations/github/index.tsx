"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import {
  Badge,
  Box,
  Button,
  Command,
  Dialog,
  Divider,
  Flex,
  Menu,
  Popover,
  Switch,
  Text,
} from "ui";
import { CheckIcon, GitIcon, PlusIcon } from "icons";
import { TeamColor } from "@/components/ui/team-color";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import { useWorkspacePath } from "@/hooks";
import {
  useCreateGitHubInstallSession,
  useCreateGitHubIssueSyncLink,
  useDeleteGitHubIssueSyncLink,
  useGitHubIntegration,
  useResyncGitHubRepositories,
  useUpdateGitHubWorkspaceSettings,
} from "@/lib/hooks/github";
import { GITHUB_BRANCH_FORMATS } from "./branch-format";

const GitHubIcon = ({ className = "h-5 w-5" }: { className?: string }) => (
  <svg
    className={className}
    fill="currentColor"
    viewBox="0 0 24 24"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
  </svg>
);

export const GitHubIntegrationSettings = () => {
  const { data: integration } = useGitHubIntegration();
  const { data: teams = [] } = useTeams();
  const { withWorkspace } = useWorkspacePath();
  const createInstallSession = useCreateGitHubInstallSession();
  const resyncRepositories = useResyncGitHubRepositories();
  const createLink = useCreateGitHubIssueSyncLink();
  const deleteLink = useDeleteGitHubIssueSyncLink();
  const updateSettings = useUpdateGitHubWorkspaceSettings();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [repositoryId, setRepositoryId] = useState("");
  const [teamId, setTeamId] = useState("");
  const [syncDirection, setSyncDirection] = useState<
    "inbound_only" | "bidirectional"
  >("inbound_only");
  const [branchFormatOpen, setBranchFormatOpen] = useState(false);
  const [repoPickerOpen, setRepoPickerOpen] = useState(false);
  const [teamPickerOpen, setTeamPickerOpen] = useState(false);

  const availableRepositories = useMemo(
    () =>
      (integration?.repositories ?? []).filter(
        (repository) =>
          !integration?.issueSyncLinks.some(
            (link) => link.repositoryId === repository.id,
          ),
      ),
    [integration],
  );

  const selectedRepo = availableRepositories.find((r) => r.id === repositoryId);
  const selectedTeam = teams.find((t) => t.id === teamId);

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        GitHub
      </Text>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Flex gap={2}>
              <Button
                color="tertiary"
                onClick={() => {
                  resyncRepositories.mutate();
                }}
              >
                Resync
              </Button>
              <Button
                color="invert"
                onClick={() => {
                  createInstallSession.mutate();
                }}
              >
                Connect GitHub
              </Button>
            </Flex>
          }
          description="Connect GitHub organizations, sync repositories, and automate team workflows."
          title="Connected organizations"
        />

        {(integration?.installations.length ?? 0) === 0 ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">
              No GitHub organizations connected
            </Text>
            <Text className="mt-1" color="muted">
              Install the FortyOne GitHub App to sync pull requests, branches,
              commits, and issues.
            </Text>
          </Box>
        ) : (
          <Box className="divide-border divide-y">
            {integration?.installations.map((installation) => (
              <Flex
                align="center"
                className="px-6 py-4"
                justify="between"
                key={installation.id}
              >
                <Flex align="center" gap={3}>
                  <Flex
                    align="center"
                    className="bg-surface-muted size-9 shrink-0 rounded-lg"
                    justify="center"
                  >
                    <GitHubIcon className="h-4.5 w-4.5" />
                  </Flex>
                  <Box>
                    <Text className="font-medium">
                      {installation.accountLogin}
                    </Text>
                    <Text color="muted">
                      {installation.repositorySelection} repositories
                    </Text>
                  </Box>
                </Flex>
                <Badge color="success" className="uppercase">
                  {installation.isActive ? "Connected" : "Disconnected"}
                </Badge>
              </Flex>
            ))}
          </Box>
        )}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          action={
            <Button
              color="tertiary"
              leftIcon={<PlusIcon />}
              onClick={() => {
                setIsModalOpen(true);
              }}
            >
              Link repository
            </Button>
          }
          description="Choose which repository syncs GitHub issues into which team."
          title="GitHub Issues"
        />
        {(integration?.issueSyncLinks.length ?? 0) === 0 ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No repositories linked to teams</Text>
            <Text className="mt-1" color="muted">
              Link a repository to enable issue sync and pull request
              automation.
            </Text>
          </Box>
        ) : (
          <Box className="divide-border divide-y">
            {integration?.issueSyncLinks.map((link) => (
              <Flex
                align="center"
                className="px-6 py-4"
                justify="between"
                key={link.id}
              >
                <Flex align="center" gap={3}>
                  <Flex
                    align="center"
                    className="bg-surface-muted size-9 shrink-0 rounded-lg"
                    justify="center"
                  >
                    <GitIcon className="h-4" />
                  </Flex>
                  <Box>
                    <Text className="font-medium">{link.repositoryName}</Text>
                    <Text color="muted">
                      {link.teamName} ·{" "}
                      {link.syncDirection === "bidirectional"
                        ? "Two-way sync"
                        : "GitHub to FortyOne only"}
                    </Text>
                  </Box>
                </Flex>
                <Button
                  color="tertiary"
                  onClick={() => {
                    deleteLink.mutate(link.id);
                  }}
                >
                  Unlink
                </Button>
              </Flex>
            ))}
          </Box>
        )}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="Keep branch names consistent across the workspace."
          title="Branch format"
        />
        <Flex align="center" className="px-6 py-4" justify="between">
          <Box>
            <Text className="font-medium">Format</Text>
            <Text color="muted">
              Choose the branch name template copied from story pages.
            </Text>
          </Box>
          <Menu open={branchFormatOpen} onOpenChange={setBranchFormatOpen}>
            <Menu.Button>
              <Button color="tertiary">
                {integration?.settings.branchFormat ??
                  "username/identifier-title"}
              </Button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                {GITHUB_BRANCH_FORMATS.map((format) => (
                  <Menu.Item
                    key={format}
                    onSelect={() => {
                      updateSettings.mutate({
                        branchFormat: format,
                      });
                    }}
                  >
                    {format}
                  </Menu.Item>
                ))}
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="Include magic words in commit messages to link commits to stories."
          title="Link commits to stories with magic words"
        />
        <Flex align="center" className="px-6 py-4" justify="between">
          <Box>
            <Text className="font-medium">Enable commit linking</Text>
            <Text color="muted">
              Example: Fixes PROD-22 or Part of PROD-22.
            </Text>
          </Box>
          <Switch
            checked={integration?.settings.linkCommitsByMagicWords ?? true}
            onCheckedChange={(checked) => {
              updateSettings.mutate({ linkCommitsByMagicWords: checked });
            }}
          />
        </Flex>
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="Control how GitHub activity syncs with FortyOne."
          title="Sync settings"
        />
        <Box className="divide-border divide-y-[0.5px]">
          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">Close stories on commit keywords</Text>
              <Text color="muted">
                Detect keywords like &quot;fixes PRO-123&quot; in commits to
                auto-close stories.
              </Text>
            </Box>
            <Switch
              checked={integration?.settings.closeOnCommitKeywords ?? true}
              onCheckedChange={(checked) => {
                updateSettings.mutate({ closeOnCommitKeywords: checked });
              }}
            />
          </Flex>
          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">Auto-populate PR body</Text>
              <Text color="muted">
                Pre-fill pull request descriptions with linked story details.
              </Text>
            </Box>
            <Switch
              checked={integration?.settings.autoPopulatePrBody ?? true}
              onCheckedChange={(checked) => {
                updateSettings.mutate({ autoPopulatePrBody: checked });
              }}
            />
          </Flex>
          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">Sync assignees</Text>
              <Text color="muted">
                Assign stories to the linked FortyOne user when a PR is opened.
              </Text>
            </Box>
            <Switch
              checked={integration?.settings.syncAssignees ?? false}
              onCheckedChange={(checked) => {
                updateSettings.mutate({ syncAssignees: checked });
              }}
            />
          </Flex>
          <Flex align="center" className="px-6 py-4" justify="between">
            <Box>
              <Text className="font-medium">Sync labels</Text>
              <Text color="muted">
                Mirror GitHub labels to FortyOne story labels.
              </Text>
            </Box>
            <Switch
              checked={integration?.settings.syncLabels ?? false}
              onCheckedChange={(checked) => {
                updateSettings.mutate({ syncLabels: checked });
              }}
            />
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6">
        <Link href={withWorkspace("/settings/workspace/integrations")}>
          <Text color="muted">Back to integrations</Text>
        </Link>
      </Box>

      <Dialog onOpenChange={setIsModalOpen} open={isModalOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Link GitHub repo to FortyOne team
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-5">
            <Box>
              <Text className="mb-2 font-medium">GitHub repository</Text>
              <Popover onOpenChange={setRepoPickerOpen} open={repoPickerOpen}>
                <Popover.Trigger asChild>
                  <Button className="w-full justify-start" color="tertiary">
                    <Flex align="center" gap={2}>
                      <GitHubIcon className="h-4 w-4 shrink-0" />
                      <Text>
                        {selectedRepo?.fullName ?? "Choose repository..."}
                      </Text>
                    </Flex>
                  </Button>
                </Popover.Trigger>
                <Popover.Content align="start" className="w-80">
                  <Command>
                    <Command.Input
                      autoFocus
                      placeholder="Search repositories..."
                    />
                    <Divider className="my-2" />
                    <Command.Empty className="py-2">
                      <Text color="muted">No repositories found.</Text>
                    </Command.Empty>
                    <Command.Group className="max-h-60 overflow-y-auto">
                      {availableRepositories.map((repository) => (
                        <Command.Item
                          active={repositoryId === repository.id}
                          className="justify-between"
                          key={repository.id}
                          onSelect={() => {
                            setRepositoryId(repository.id);
                            setRepoPickerOpen(false);
                          }}
                        >
                          <Flex align="center" gap={2}>
                            <GitHubIcon className="h-4 w-4 shrink-0" />
                            <Text className="truncate">
                              {repository.fullName}
                            </Text>
                          </Flex>
                          {repositoryId === repository.id && (
                            <CheckIcon
                              className="h-5 w-auto shrink-0"
                              strokeWidth={2.1}
                            />
                          )}
                        </Command.Item>
                      ))}
                    </Command.Group>
                  </Command>
                </Popover.Content>
              </Popover>
            </Box>

            <Box>
              <Text className="mb-2 font-medium">FortyOne team</Text>
              <Popover onOpenChange={setTeamPickerOpen} open={teamPickerOpen}>
                <Popover.Trigger asChild>
                  <Button className="w-full justify-start" color="tertiary">
                    <Flex align="center" gap={2}>
                      <TeamColor
                        className="shrink-0"
                        color={selectedTeam?.color}
                      />
                      <Text>{selectedTeam?.name ?? "Choose team..."}</Text>
                    </Flex>
                  </Button>
                </Popover.Trigger>
                <Popover.Content align="start" className="w-80">
                  <Command>
                    <Command.Input autoFocus placeholder="Search teams..." />
                    <Divider className="my-2" />
                    <Command.Empty className="py-2">
                      <Text color="muted">No teams found.</Text>
                    </Command.Empty>
                    <Command.Group className="max-h-60 overflow-y-auto">
                      {teams.map((team) => (
                        <Command.Item
                          active={teamId === team.id}
                          className="justify-between"
                          key={team.id}
                          onSelect={() => {
                            setTeamId(team.id);
                            setTeamPickerOpen(false);
                          }}
                        >
                          <Flex align="center" gap={2}>
                            <TeamColor
                              className="shrink-0"
                              color={team.color}
                            />
                            <Text className="truncate">{team.name}</Text>
                          </Flex>
                          {teamId === team.id && (
                            <CheckIcon
                              className="h-5 w-auto shrink-0"
                              strokeWidth={2.1}
                            />
                          )}
                        </Command.Item>
                      ))}
                    </Command.Group>
                  </Command>
                </Popover.Content>
              </Popover>
            </Box>

            <Box>
              <Text className="mb-2 font-medium">Sync mode</Text>
              <Flex direction="column" gap={2}>
                <button
                  className={`border-border rounded-lg border px-4 py-3 text-left transition-colors ${syncDirection === "inbound_only" ? "border-primary bg-primary/5" : "hover:bg-surface-muted"}`}
                  onClick={() => setSyncDirection("inbound_only")}
                  type="button"
                >
                  <Text className="font-medium">One-way sync</Text>
                  <Text color="muted">
                    Only sync issues from GitHub to FortyOne
                  </Text>
                </button>
                <button
                  className={`border-border rounded-lg border px-4 py-3 text-left transition-colors ${syncDirection === "bidirectional" ? "border-primary bg-primary/5" : "hover:bg-surface-muted"}`}
                  onClick={() => setSyncDirection("bidirectional")}
                  type="button"
                >
                  <Text className="font-medium">Two-way sync</Text>
                  <Text color="muted">
                    Sync issues between FortyOne and GitHub
                  </Text>
                </button>
              </Flex>
            </Box>

            <Flex justify="end" gap={2}>
              <Button color="tertiary" onClick={() => setIsModalOpen(false)}>
                Cancel
              </Button>
              <Button
                disabled={!repositoryId || !teamId}
                onClick={() => {
                  createLink.mutate(
                    { repositoryId, teamId, syncDirection },
                    {
                      onSuccess: () => {
                        setIsModalOpen(false);
                        setRepositoryId("");
                        setTeamId("");
                        setSyncDirection("inbound_only");
                      },
                    },
                  );
                }}
              >
                Save
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
