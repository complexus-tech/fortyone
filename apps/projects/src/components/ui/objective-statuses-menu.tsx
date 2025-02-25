"use client";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon } from "icons";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { ObjectiveStatusIcon } from "./objective-status-icon";

const StatusContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useStatusMenu = () => {
  const { open, setOpen } = useContext(StatusContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useStatusMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const ObjectiveStatusesMenu = ({
  children,
}: {
  children: ReactNode;
}) => {
  const [open, setOpen] = useState(false);
  return (
    <StatusContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </StatusContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  statusId,
  setStatusId,
}: {
  statusId?: string;
  setStatusId: (statusId: string) => void;
}) => {
  const { data: statuses = [] } = useObjectiveStatuses();
  const [query, setQuery] = useState("");

  const state =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const { id: defaultStateId } = state!;
  const { setOpen } = useStatusMenu();
  if (!statuses.length) return null;

  return (
    <Popover.Content align="center" className="w-64">
      <Command>
        <Command.Input
          autoFocus
          onValueChange={(value) => {
            if (Number.parseInt(value) < statuses.length) {
              setStatusId(statuses[Number.parseInt(value)].id);
              setOpen(false);
              setQuery("");
              return;
            }
            setQuery(value);
          }}
          placeholder="Change status..."
          value={query}
        />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No statuses found.</Text>
        </Command.Empty>
        <Command.Group>
          {statuses.map(({ id, name }, idx) => (
            <Command.Item
              active={id === defaultStateId}
              className="justify-between"
              key={id}
              onSelect={() => {
                if (id !== defaultStateId) {
                  setStatusId(id);
                }
                setOpen(false);
              }}
              value={name}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <ObjectiveStatusIcon statusId={id} />
                <Text>{name}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {id === defaultStateId && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

ObjectiveStatusesMenu.Trigger = Trigger;
ObjectiveStatusesMenu.Items = Items;
