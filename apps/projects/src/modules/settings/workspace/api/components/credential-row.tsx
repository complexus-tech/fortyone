"use client";

import { DeleteIcon, MoreHorizontalIcon, RefreshIcon } from "icons";
import { Box, Button, Flex, Menu, Text } from "ui";
import { formatDeveloperDate } from "../constants";
import type { Credential } from "../types";

export const CredentialRow = ({
  credential,
  isPending,
  onRevoke,
  onRotate,
}: {
  credential: Credential;
  isPending: boolean;
  onRevoke: () => void;
  onRotate: () => void;
}) => {
  const isRevoked = Boolean(credential.revokedAt);
  return (
    <Flex align="center" className="gap-4 px-6 py-4" justify="between" wrap>
      <Box className="min-w-0 flex-1">
        <Flex align="center" gap={2}>
          <Text className="font-medium">{credential.name}</Text>
          {isRevoked ? <Text color="muted">Revoked</Text> : null}
        </Flex>
        <Text className="mt-1 break-all" color="muted">
          Prefix {credential.prefix} · Expires{" "}
          {formatDeveloperDate(credential.expiresAt)}
          {credential.lastUsedAt
            ? ` · Last used ${formatDeveloperDate(credential.lastUsedAt)}`
            : " · Never used"}
        </Text>
        <Text className="mt-1" color="muted">
          {credential.scopes.join(", ")}
        </Text>
      </Box>
      {!isRevoked ? (
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${credential.name}`}
              asIcon
              color="tertiary"
              disabled={isPending}
              leftIcon={<MoreHorizontalIcon />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">More options</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item disabled={isPending} onSelect={onRotate}>
                <RefreshIcon className="h-[1.15rem]" />
                Rotate
              </Menu.Item>
              <Menu.Item disabled={isPending} onSelect={onRevoke}>
                <DeleteIcon className="h-[1.15rem]" />
                Revoke
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      ) : null}
    </Flex>
  );
};
