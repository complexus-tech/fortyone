"use client";
import {
  createContext,
  use,
  useDeferredValue,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import { Box, Command, Divider, Flex, Popover, Text } from "ui";
import { CheckIcon, ObjectiveIcon } from "icons";
import {
  OBJECTIVE_MENU_PAGE_SIZE,
  useTeamObjectivesInfinite,
} from "@/modules/objectives/hooks/use-objectives";
import { useTerminology } from "@/hooks";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";

const ObjectivesContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,

  setOpen: () => {},
});

export const useObjectivesMenu = () => {
  const { open, setOpen } = use(ObjectivesContext);
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
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { open, setOpen } = useObjectivesMenu();
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending } =
    useTeamObjectivesInfinite(
      teamId ?? "",
      deferredQuery,
      OBJECTIVE_MENU_PAGE_SIZE,
      open,
    );
  const objectives = data?.pages.flatMap((page) => page.objectives) ?? [];

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom <= 80 && hasNextPage && !isFetchingNextPage) {
      void fetchNextPage();
    }
  };

  return (
    <Popover.Content align={align}>
      <Command>
        <Command.Input
          autoFocus
          onValueChange={(value) => {
            const itemIndex = Number.parseInt(value, 10);
            if (/^\d+$/.test(value) && itemIndex < objectives.length) {
              setObjectiveId(objectives[itemIndex].id);
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
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
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
              <MenuLoadingSkeleton rows={5} />
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
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

ObjectivesMenu.Trigger = Trigger;
ObjectivesMenu.Items = Items;
