import type { ReactNode } from "react";
import { Box, Flex, Menu, Text } from "ui";
import { CheckIcon } from "icons";
import { StoryStatusIcon } from "../story-status-icon";
import { useStore } from "@/hooks/store";

export const StatusesMenu = ({ children }: { children: ReactNode }) => {
  return <Menu>{children}</Menu>;
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Menu.Button>{children}</Menu.Button>
);

const Items = ({
  statusId,
  isSearchEnabled = true,
  setStatusId,
}: {
  statusId?: string;
  isSearchEnabled?: boolean;
  setStatusId: (statusId: string) => void;
}) => {
  const { states } = useStore();
  if (!states.length) return null;
  const state = states.find((state) => state.id === statusId) || states.at(0);
  const { id: defaultStateId } = state!!;

  return (
    <Menu.Items align="center" className="w-64">
      {isSearchEnabled ? (
        <>
          <Menu.Group className="px-4">
            <Menu.Input autoFocus placeholder="Change status..." />
          </Menu.Group>
          <Menu.Separator className="my-2" />
        </>
      ) : null}
      <Menu.Group>
        {states.map(({ id, name }, idx) => (
          <Menu.Item
            active={id === defaultStateId}
            onClick={() => {
              setStatusId(id);
            }}
            className="justify-between"
            key={id}
          >
            <Box className="grid grid-cols-[24px_auto] items-center">
              <StoryStatusIcon statusId={id} />
              <Text>{name}</Text>
            </Box>
            <Flex align="center" gap={2}>
              {id === defaultStateId && (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              )}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Menu.Item>
        ))}
      </Menu.Group>
    </Menu.Items>
  );
};

StatusesMenu.Trigger = Trigger;
StatusesMenu.Items = Items;
