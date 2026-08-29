"use client";

import { useState } from "react";
import { Box, Button, Checkbox, Dialog, Flex, Input, Text } from "ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { DEVELOPER_SCOPES, expiryFromNow } from "../constants";
import type { CreateCredentialInput, DeveloperScope } from "../types";

type CredentialDialogProps = {
  description: string;
  isPending: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: CreateCredentialInput) => Promise<void>;
  open: boolean;
  title: string;
};

const DEFAULT_SCOPES: DeveloperScope[] = ["stories:read"];

export const CredentialDialog = ({
  description,
  isPending,
  onOpenChange,
  onSubmit,
  open,
  title,
}: CredentialDialogProps) => {
  const { data: teams = [] } = useTeams();
  const [name, setName] = useState("");
  const [expiryDays, setExpiryDays] = useState("90");
  const [scopes, setScopes] = useState<DeveloperScope[]>(DEFAULT_SCOPES);
  const [teamIds, setTeamIds] = useState<string[]>([]);
  const selectedScopes = new Set(scopes);
  const selectedTeamIds = new Set(teamIds);

  const reset = () => {
    setName("");
    setExpiryDays("90");
    setScopes(DEFAULT_SCOPES);
    setTeamIds([]);
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && !isPending) reset();
    if (!isPending) onOpenChange(nextOpen);
  };

  const toggleScope = (scope: DeveloperScope) => {
    setScopes((current) =>
      current.includes(scope)
        ? current.filter((value) => value !== scope)
        : [...current, scope],
    );
  };

  const toggleTeam = (teamId: string) => {
    setTeamIds((current) =>
      current.includes(teamId)
        ? current.filter((value) => value !== teamId)
        : [...current, teamId],
    );
  };

  const days = Number(expiryDays);
  const isValid =
    name.trim().length > 0 &&
    scopes.length > 0 &&
    Number.isInteger(days) &&
    days >= 1 &&
    days <= 365;

  const submit = async () => {
    try {
      await onSubmit({
        name: name.trim(),
        scopes,
        teamIds,
        expiresAt: expiryFromNow(days),
      });
      reset();
    } catch {
      // The owning section reports the API error and keeps the draft intact.
    }
  };

  return (
    <Dialog onOpenChange={handleOpenChange} open={open}>
      <Dialog.Content
        className="mt-0 max-w-3xl md:mt-0"
        overlayClassName="items-center py-6"
        size="lg"
      >
        <Dialog.Header className="px-6 pt-6 pb-2">
          <Dialog.Title className="text-xl">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-5 px-6 pt-2 pb-4">
          <Text color="muted">{description}</Text>
          <Input
            autoFocus
            label="Name"
            maxLength={120}
            onChange={(event) => {
              setName(event.target.value);
            }}
            placeholder="Production automation"
            value={name}
          />
          <Box>
            <Text className="mb-2" fontWeight="medium">
              Permissions
            </Text>
            <Box className="grid max-h-64 grid-cols-1 gap-2 overflow-y-auto sm:grid-cols-2 lg:grid-cols-3">
              {DEVELOPER_SCOPES.map((scope) => (
                <label
                  className="border-border hover:bg-state-hover flex cursor-pointer items-start gap-3 rounded-xl border px-3 py-2.5"
                  htmlFor={`developer-scope-${scope.value}`}
                  key={scope.value}
                >
                  <Checkbox
                    checked={selectedScopes.has(scope.value)}
                    className="relative top-0.5"
                    id={`developer-scope-${scope.value}`}
                    onCheckedChange={() => {
                      toggleScope(scope.value);
                    }}
                  />
                  <Text fontWeight="medium">{scope.label}</Text>
                </label>
              ))}
            </Box>
          </Box>
          {teams.length > 0 ? (
            <Box>
              <Text className="mb-1" fontWeight="medium">
                Team restrictions
              </Text>
              <Text className="mb-2" color="muted">
                Leave every team unchecked for workspace-wide access.
              </Text>
              <Flex className="max-h-32 flex-wrap overflow-y-auto" gap={2}>
                {teams.map((team) => (
                  <label
                    className="border-border hover:bg-state-hover flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2"
                    htmlFor={`developer-team-${team.id}`}
                    key={team.id}
                  >
                    <Checkbox
                      checked={selectedTeamIds.has(team.id)}
                      id={`developer-team-${team.id}`}
                      onCheckedChange={() => {
                        toggleTeam(team.id);
                      }}
                    />
                    <Text>{team.name}</Text>
                  </label>
                ))}
              </Flex>
            </Box>
          ) : null}
          <Box className="max-w-52">
            <Text className="mb-[0.35rem]" fontWeight="medium">
              Expires after
            </Text>
            <Flex align="center" gap={2}>
              <Input
                aria-label="Expiration in days"
                max={365}
                min={1}
                onChange={(event) => {
                  setExpiryDays(event.target.value);
                }}
                type="number"
                value={expiryDays}
              />
              <Text color="muted">days</Text>
            </Flex>
          </Box>
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-2">
          <Button
            color="tertiary"
            disabled={isPending}
            onClick={() => {
              handleOpenChange(false);
            }}
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            disabled={!isValid}
            loading={isPending}
            onClick={() => void submit()}
          >
            Create
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
