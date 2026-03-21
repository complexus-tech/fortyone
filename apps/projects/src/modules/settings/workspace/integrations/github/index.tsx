"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Box, Button, Dialog, Flex, Select, Switch, Text } from "ui";
import { GitIcon, PlusIcon, ReloadIcon } from "icons";
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

  return (
    <Box>
      <Text
        as="h1"
        className="mb-6 flex items-center gap-2 text-2xl font-medium"
      >
        <GitIcon className="h-5" />
        GitHub
      </Text>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Flex gap={2}>
              <Button
                color="secondary"
                leftIcon={<ReloadIcon />}
                onClick={() => {
                  resyncRepositories.mutate();
                }}
              >
                Resync
              </Button>
              <Button
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
                <Box>
                  <Text className="font-medium">
                    {installation.accountLogin}
                  </Text>
                  <Text color="muted">
                    {installation.repositorySelection} repositories
                  </Text>
                </Box>
                <Text color="muted">
                  {installation.isActive ? "Connected" : "Disconnected"}
                </Text>
              </Flex>
            ))}
          </Box>
        )}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          action={
            <Button
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
                <Box>
                  <Text className="font-medium">{link.repositoryName}</Text>
                  <Text color="muted">
                    {link.teamName} ·{" "}
                    {link.syncDirection === "bidirectional"
                      ? "Two-way sync"
                      : "GitHub to FortyOne only"}
                  </Text>
                </Box>
                <Button
                  color="secondary"
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
              Current story branch names default to username, identifier, and
              title.
            </Text>
          </Box>
          <Select
            onValueChange={(value) => {
              updateSettings.mutate({ branchFormat: value });
            }}
            value={
              integration?.settings.branchFormat ?? "username/identifier-title"
            }
          >
            <Select.Trigger className="w-64 text-[0.9rem] md:text-base">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option value="username/identifier-title">
                username/identifier-title
              </Select.Option>
            </Select.Content>
          </Select>
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
              <Select onValueChange={setRepositoryId} value={repositoryId}>
                <Select.Trigger className="w-full text-[0.9rem] md:text-base">
                  <Select.Input placeholder="Choose repository..." />
                </Select.Trigger>
                <Select.Content>
                  {availableRepositories.map((repository) => (
                    <Select.Option key={repository.id} value={repository.id}>
                      {repository.fullName}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Box>

            <Box>
              <Text className="mb-2 font-medium">FortyOne team</Text>
              <Select onValueChange={setTeamId} value={teamId}>
                <Select.Trigger className="w-full text-[0.9rem] md:text-base">
                  <Select.Input placeholder="Choose team..." />
                </Select.Trigger>
                <Select.Content>
                  {teams.map((team) => (
                    <Select.Option key={team.id} value={team.id}>
                      {team.name}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Box>

            <Box>
              <Text className="mb-2 font-medium">Sync mode</Text>
              <Select
                onValueChange={(value) => {
                  setSyncDirection(value as "inbound_only" | "bidirectional");
                }}
                value={syncDirection}
              >
                <Select.Trigger className="w-full text-[0.9rem] md:text-base">
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  <Select.Option value="inbound_only">
                    Only sync issues from this repository to FortyOne
                  </Select.Option>
                  <Select.Option value="bidirectional">
                    Two-way sync issues between FortyOne and this repository
                  </Select.Option>
                </Select.Content>
              </Select>
            </Box>

            <Flex justify="end">
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
