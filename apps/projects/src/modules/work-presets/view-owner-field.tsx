"use client";
import { useQuery } from "@tanstack/react-query";
import { Button, Flex, Select, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getJoinedTeams } from "@/modules/teams/public/queries";
import { teamKeys } from "@/constants/keys";

export const ViewOwnerField = ({
  value,
  onChange,
  disabled,
}: {
  value: string;
  onChange: (teamId: string) => void;
  disabled: boolean;
}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const teams = useQuery({
    queryKey: [
      ...teamKeys.joined(workspaceSlug),
      "view-owner",
      session?.user.id,
    ],
    queryFn: () => getJoinedTeams({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
    refetchOnMount: "always",
  });
  return (
    <div className="space-y-2">
      <Flex align="center" gap={3} justify="between">
        <Text id="view-owner-label">Save to team</Text>
        <Select
          disabled={disabled || teams.isPending || teams.isError}
          onValueChange={(next: string) => {
            if (!disabled) onChange(next);
          }}
          value={value}
        >
          <Select.Trigger
            aria-labelledby="view-owner-label"
            className="max-w-64"
          >
            <Select.Input
              placeholder={
                teams.isPending ? "Loading teams..." : "Choose a team"
              }
            >
              {teams.data?.find((team) => team.id === value)?.name}
            </Select.Input>
          </Select.Trigger>
          <Select.Content>
            {teams.data?.map((team) => (
              <Select.Option key={team.id} value={team.id}>
                {team.name}
              </Select.Option>
            ))}
          </Select.Content>
        </Select>
      </Flex>
      {teams.isError ? (
        <Flex align="center" gap={2} justify="between">
          <Text color="danger" role="alert">
            Teams could not be loaded.
          </Text>
          <Button
            color="tertiary"
            disabled={disabled || teams.isFetching}
            onClick={() => {
              void teams.refetch();
            }}
            size="sm"
            type="button"
            variant="outline"
          >
            Try again
          </Button>
        </Flex>
      ) : null}
      {!teams.isPending && !teams.isError && !teams.data.length ? (
        <Text color="muted">Join a team before creating a saved view.</Text>
      ) : null}
    </div>
  );
};
