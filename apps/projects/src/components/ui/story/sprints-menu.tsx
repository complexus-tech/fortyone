"use client";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, LoadingIcon, SprintsIcon } from "icons";
import { format } from "date-fns";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";

const SprintsContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useSprintsMenu = () => {
  const { open, setOpen } = useContext(SprintsContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useSprintsMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const SprintsMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <SprintsContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </SprintsContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  align = "center",
  sprintId,
  setSprintId,
  teamId,
}: {
  sprintId?: string;
  setSprintId: (sprintId: string | null) => void;
  align?: "center" | "start" | "end" | undefined;
  teamId?: string;
  objectiveId?: string;
}) => {
  const { data: sprints = [], isPending: isTeamSprintsPending } =
    useTeamSprints(teamId ?? "");
  const [query, setQuery] = useState("");
  const { setOpen } = useSprintsMenu();

  return (
    <Popover.Content align={align}>
      <Command>
        <Command.Input
          autoFocus
          onValueChange={(value) => {
            if (Number.parseInt(value) < sprints.length) {
              setSprintId(sprints[Number.parseInt(value)].id);
              setOpen(false);
              setQuery("");
              return;
            }
            setQuery(value);
          }}
          placeholder="Add to sprint..."
          value={query}
        />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No sprint found.</Text>
        </Command.Empty>
        <Command.Group>
          {!isTeamSprintsPending ? (
            <Command.Item
              active={!sprintId}
              className="justify-between gap-4 opacity-70"
              onSelect={() => {
                if (sprintId) {
                  setSprintId(null);
                }
                setOpen(false);
              }}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <SprintsIcon />
                <Text>No sprint</Text>
              </Box>
              <Flex align="center" gap={1}>
                {!sprintId && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">0</Text>
              </Flex>
            </Command.Item>
          ) : null}
          {sprints.length > 0 && <Divider className="my-2" />}
          {isTeamSprintsPending ? (
            <Command.Loading className="p-2">
              <Text className="flex items-center gap-2" color="muted">
                <LoadingIcon className="animate-spin" />
                Fetching sprints…
              </Text>
            </Command.Loading>
          ) : null}
          {sprints.map(({ id, name, startDate, endDate }, idx) => (
            <Command.Item
              active={id === sprintId}
              className="justify-between gap-4"
              key={id}
              onSelect={() => {
                if (id !== sprintId) {
                  setSprintId(id);
                }
                setOpen(false);
              }}
              value={name}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <SprintsIcon className="h-5 w-auto" />
                <Text className="mr-4 flex items-center gap-3">
                  {name}
                  <Text as="span" color="muted">
                    {format(new Date(startDate), "MMM d")} -{" "}
                    {format(new Date(endDate), "MMM d")}
                  </Text>
                </Text>
              </Box>
              <Flex align="center" gap={1}>
                {id === sprintId && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx + 1}</Text>
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

SprintsMenu.Trigger = Trigger;
SprintsMenu.Items = Items;
