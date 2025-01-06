"use client";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, ObjectiveIcon } from "icons";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";

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
    <Popover open={open} onOpenChange={setOpen}>
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
}: {
  objectiveId?: string;
  setObjectiveId: (objectiveId: string | null) => void;
  align?: "center" | "start" | "end" | undefined;
}) => {
  const { data: objectives = [] } = useObjectives();
  const [query, setQuery] = useState("");
  if (!objectives.length) return null;
  const { setOpen } = useObjectivesMenu();

  return (
    <Popover.Content align={align}>
      <Command>
        <Command.Input
          autoFocus
          placeholder="Change objective..."
          value={query}
          onValueChange={(value) => {
            if (Number.parseInt(value) < objectives.length) {
              setObjectiveId(objectives[Number.parseInt(value)].id);
              setOpen(false);
              setQuery("");
              return;
            }
            setQuery(value);
          }}
        />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No objective found.</Text>
        </Command.Empty>
        <Command.Group>
          <Command.Item
            active={!objectiveId}
            onSelect={() => {
              if (objectiveId) {
                setObjectiveId(null);
              }
              setOpen(false);
            }}
            className="justify-between gap-4"
          >
            <Box className="grid grid-cols-[24px_auto] items-center">
              <ObjectiveIcon className="h-[1.1rem]" />
              <Text>No objective</Text>
            </Box>
            <Flex align="center" gap={1}>
              {!objectiveId && (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              )}
              <Text color="muted">0</Text>
            </Flex>
          </Command.Item>
          {objectives.map(({ id, name }, idx) => (
            <Command.Item
              active={id === objectiveId}
              value={name}
              onSelect={() => {
                if (id !== objectiveId) {
                  setObjectiveId(id);
                }
                setOpen(false);
              }}
              className="justify-between gap-4"
              key={id}
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
