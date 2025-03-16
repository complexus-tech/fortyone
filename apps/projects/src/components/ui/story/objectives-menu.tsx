"use client";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, ObjectiveIcon, LoadingIcon } from "icons";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTerminology } from "@/hooks";

const ObjectivesContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,

  setOpen: () => {},
});

export const useObjectivesMenu = () => {
  const { open, setOpen } = useContext(ObjectivesContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useObjectivesMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const ObjectivesMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <ObjectivesContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </ObjectivesContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  align = "center",
  objectiveId,
  setObjectiveId,
  teamId,
}: {
  objectiveId?: string;
  setObjectiveId: (objectiveId: string | null) => void;
  align?: "center" | "start" | "end" | undefined;
  teamId?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const { data: objectives = [], isPending } = useTeamObjectives(teamId ?? "");
  const [query, setQuery] = useState("");
  const { setOpen } = useObjectivesMenu();

  return (
    <Popover.Content align={align}>
      <Command>
        <Command.Input
          autoFocus
          onValueChange={(value) => {
            if (Number.parseInt(value) < objectives.length) {
              setObjectiveId(objectives[Number.parseInt(value)].id);
              setOpen(false);
              setQuery("");
              return;
            }
            setQuery(value);
          }}
          placeholder={`Change ${getTermDisplay("objectiveTerm")}...`}
          value={query}
        />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No {getTermDisplay("objectiveTerm")} found.</Text>
        </Command.Empty>
        <Command.Group>
          {!isPending && (
            <Command.Item
              active={!objectiveId}
              className="justify-between gap-4 opacity-70"
              onSelect={() => {
                if (objectiveId) {
                  setObjectiveId(null);
                }
                setOpen(false);
              }}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <ObjectiveIcon className="h-[1.1rem]" />
                <Text>No {getTermDisplay("objectiveTerm")}</Text>
              </Box>
              <Flex align="center" gap={1}>
                {!objectiveId && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">0</Text>
              </Flex>
            </Command.Item>
          )}
          {objectives.length > 0 && <Divider className="my-2" />}
          {isPending ? (
            <Command.Loading className="p-2">
              <Text className="flex items-center gap-2" color="muted">
                <LoadingIcon className="animate-spin" />
                Fetching {getTermDisplay("objectiveTerm")}…
              </Text>
            </Command.Loading>
          ) : null}
          {objectives.map(({ id, name }, idx) => (
            <Command.Item
              active={id === objectiveId}
              className="justify-between gap-4"
              key={id}
              onSelect={() => {
                if (id !== objectiveId) {
                  setObjectiveId(id);
                }
                setOpen(false);
              }}
              value={name}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <ObjectiveIcon className="h-[1.1rem]" />
                <Text>{name}</Text>
              </Box>
              <Flex align="center" gap={1}>
                {id === objectiveId && (
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

ObjectivesMenu.Trigger = Trigger;
ObjectivesMenu.Items = Items;
