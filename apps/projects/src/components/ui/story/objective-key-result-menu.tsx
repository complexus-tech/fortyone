"use client";

import type { ReactNode } from "react";
import { useDeferredValue, useState } from "react";
import { ArrowRightIcon, CheckIcon, ObjectiveIcon, OKRIcon } from "icons";
import { Box, Divider, Flex, Menu, Text } from "ui";
import { useTerminology } from "@/hooks";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import type { Objective } from "@/modules/objectives/types";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";
import { filterKeyResultsByName } from "./objective-key-result-menu-utils";

type ObjectiveKeyResultSelection = {
  objectiveId: string | null;
  keyResultId: string | null;
};

type ObjectiveSubMenuProps = {
  currentKeyResultId: string | null;
  currentObjectiveId: string | null;
  objective: Objective;
  onSelect: (selection: ObjectiveKeyResultSelection) => void;
};

const ObjectiveOnlyItem = ({
  currentObjectiveId,
  objective,
  onSelect,
}: Pick<
  ObjectiveSubMenuProps,
  "currentObjectiveId" | "objective" | "onSelect"
>) => {
  const isCurrentObjective = currentObjectiveId === objective.id;

  return (
    <Menu.Item
      active={isCurrentObjective}
      className="justify-between gap-3"
      onSelect={() => {
        onSelect({ objectiveId: objective.id, keyResultId: null });
      }}
    >
      <Flex align="center" className="min-w-0 gap-2">
        <ObjectiveIcon className="h-[1.1rem] shrink-0" />
        <Text className="max-w-64 truncate">{objective.name}</Text>
      </Flex>
      {isCurrentObjective ? <CheckIcon className="h-4 w-4 shrink-0" /> : null}
    </Menu.Item>
  );
};

const ObjectiveSubMenu = ({
  currentKeyResultId,
  currentObjectiveId,
  objective,
  onSelect,
}: ObjectiveSubMenuProps) => {
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data: keyResults = [], isPending } = useKeyResults(
    objective.id,
    isOpen,
  );
  const isCurrentObjective = currentObjectiveId === objective.id;
  const filteredKeyResults = filterKeyResultsByName(keyResults, deferredQuery);

  return (
    <Menu.SubMenu
      onOpenChange={(open) => {
        setIsOpen(open);
        if (!open) setQuery("");
      }}
      open={isOpen}
    >
      <Menu.SubTrigger
        active={isCurrentObjective}
        className="justify-between gap-4"
      >
        <Flex align="center" className="min-w-0 gap-2">
          <ObjectiveIcon className="h-[1.1rem] shrink-0" />
          <Text className="max-w-64 truncate">{objective.name}</Text>
        </Flex>
        <Flex align="center" className="shrink-0 gap-1.5">
          {isCurrentObjective ? <CheckIcon className="h-4 w-4" /> : null}
          <ArrowRightIcon
            className="text-text-muted h-3.5 w-3.5"
            strokeWidth={2.4}
          />
        </Flex>
      </Menu.SubTrigger>
      <Menu.SubItems className="w-80 max-w-[calc(100vw-2rem)]">
        <Box className="px-3 pt-0.5 pb-1.5">
          <Menu.Input
            autoFocus
            onChange={(event) => {
              setQuery(event.target.value);
            }}
            placeholder={`Find ${getTermDisplay("keyResultTerm")}...`}
            value={query}
          />
        </Box>
        <Divider className="my-1.5" />
        <Menu.Group>
          <Menu.Item
            active={Boolean(isCurrentObjective && !currentKeyResultId)}
            className="justify-between gap-3"
            onSelect={() => {
              onSelect({ objectiveId: objective.id, keyResultId: null });
            }}
          >
            <Flex align="center" className="min-w-0 gap-2">
              <ObjectiveIcon className="h-[1.1rem] shrink-0" />
              <Text className="truncate">
                Link {getTermDisplay("objectiveTerm")} only
              </Text>
            </Flex>
            {isCurrentObjective && !currentKeyResultId ? (
              <CheckIcon className="h-4 w-4 shrink-0" />
            ) : null}
          </Menu.Item>
        </Menu.Group>
        <Menu.Separator />
        <Menu.Group className="max-h-72 overflow-y-auto">
          {isPending ? <MenuLoadingSkeleton rows={3} /> : null}
          {!isPending && filteredKeyResults.length === 0 ? (
            <Text className="px-2 py-2" color="muted">
              No {getTermDisplay("keyResultTerm")} found.
            </Text>
          ) : null}
          {filteredKeyResults.map((keyResult) => (
            <Menu.Item
              active={keyResult.id === currentKeyResultId}
              className="justify-between gap-3"
              key={keyResult.id}
              onSelect={() => {
                onSelect({
                  objectiveId: objective.id,
                  keyResultId: keyResult.id,
                });
              }}
            >
              <Flex align="center" className="min-w-0 gap-2">
                <OKRIcon className="h-[1.1rem] shrink-0" strokeWidth={2.4} />
                <Text className="truncate">{keyResult.name}</Text>
              </Flex>
              {keyResult.id === currentKeyResultId ? (
                <CheckIcon className="h-4 w-4 shrink-0" />
              ) : null}
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.SubItems>
    </Menu.SubMenu>
  );
};

const ObjectiveMenuItem = (props: ObjectiveSubMenuProps) =>
  props.objective.keyResultCount > 0 ? (
    <ObjectiveSubMenu {...props} />
  ) : (
    <ObjectiveOnlyItem {...props} />
  );

export const ObjectiveKeyResultMenu = ({
  align = "center",
  children,
  keyResultId,
  objectiveId,
  onChange,
  teamId,
}: {
  align?: "center" | "end" | "start";
  children: ReactNode;
  keyResultId: string | null;
  objectiveId: string | null;
  onChange: (selection: ObjectiveKeyResultSelection) => void;
  teamId: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data: objectives = [], isPending } = useTeamObjectives(
    teamId,
    deferredQuery,
  );

  const selectAndClose = (selection: ObjectiveKeyResultSelection) => {
    onChange(selection);
    setIsOpen(false);
    setQuery("");
  };

  return (
    <Menu onOpenChange={setIsOpen} open={isOpen}>
      <Menu.Button>{children}</Menu.Button>
      <Menu.Items
        align={align}
        className="w-80 max-w-[calc(100vw-2rem)]"
        onCloseAutoFocus={() => {
          setQuery("");
        }}
      >
        <Box className="px-3 pt-0.5 pb-1.5">
          <Menu.Input
            autoFocus
            onChange={(event) => {
              setQuery(event.target.value);
            }}
            placeholder={`Find ${getTermDisplay("objectiveTerm")}...`}
            value={query}
          />
        </Box>
        <Divider className="my-1.5" />
        <Menu.Group>
          <Menu.Item
            active={!objectiveId}
            className="justify-between gap-3"
            onSelect={() => {
              selectAndClose({ objectiveId: null, keyResultId: null });
            }}
          >
            <Flex align="center" className="gap-2">
              <ObjectiveIcon className="h-[1.1rem]" />
              <Text>No {getTermDisplay("objectiveTerm")}</Text>
            </Flex>
            {!objectiveId ? <CheckIcon className="h-4 w-4" /> : null}
          </Menu.Item>
        </Menu.Group>
        <Menu.Separator />
        <Menu.Group className="max-h-80 overflow-y-auto">
          {isPending ? <MenuLoadingSkeleton rows={5} /> : null}
          {!isPending && objectives.length === 0 ? (
            <Text className="px-2 py-2" color="muted">
              No {getTermDisplay("objectiveTerm")} found.
            </Text>
          ) : null}
          {objectives.map((objective) => (
            <ObjectiveMenuItem
              currentKeyResultId={keyResultId}
              currentObjectiveId={objectiveId}
              key={objective.id}
              objective={objective}
              onSelect={selectAndClose}
            />
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

export const KeyResultMenu = ({
  align = "center",
  children,
  keyResultId,
  objectiveId,
  onChange,
}: {
  align?: "center" | "end" | "start";
  children: ReactNode;
  keyResultId: string | null;
  objectiveId: string;
  onChange: (keyResultId: string | null) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data: keyResults = [], isPending } = useKeyResults(
    objectiveId,
    isOpen,
  );
  const filteredKeyResults = filterKeyResultsByName(keyResults, deferredQuery);

  const selectAndClose = (nextKeyResultId: string | null) => {
    onChange(nextKeyResultId);
    setIsOpen(false);
    setQuery("");
  };

  return (
    <Menu
      onOpenChange={(open) => {
        setIsOpen(open);
        if (!open) setQuery("");
      }}
      open={isOpen}
    >
      <Menu.Button>{children}</Menu.Button>
      <Menu.Items align={align} className="w-80 max-w-[calc(100vw-2rem)]">
        <Box className="px-3 pt-0.5 pb-1.5">
          <Menu.Input
            autoFocus
            onChange={(event) => {
              setQuery(event.target.value);
            }}
            placeholder={`Find ${getTermDisplay("keyResultTerm")}...`}
            value={query}
          />
        </Box>
        <Divider className="my-1.5" />
        <Menu.Group>
          <Menu.Item
            active={!keyResultId}
            className="justify-between gap-3"
            onSelect={() => {
              selectAndClose(null);
            }}
          >
            <Flex align="center" className="min-w-0 gap-2">
              <OKRIcon className="h-[1.1rem] shrink-0" strokeWidth={2.4} />
              <Text className="truncate">
                No {getTermDisplay("keyResultTerm")}
              </Text>
            </Flex>
            {!keyResultId ? <CheckIcon className="h-4 w-4 shrink-0" /> : null}
          </Menu.Item>
        </Menu.Group>
        <Menu.Separator />
        <Menu.Group className="max-h-80 overflow-y-auto">
          {isPending ? <MenuLoadingSkeleton rows={4} /> : null}
          {!isPending && filteredKeyResults.length === 0 ? (
            <Text className="px-2 py-2" color="muted">
              No {getTermDisplay("keyResultTerm")} found.
            </Text>
          ) : null}
          {filteredKeyResults.map((keyResult) => (
            <Menu.Item
              active={keyResult.id === keyResultId}
              className="justify-between gap-3"
              key={keyResult.id}
              onSelect={() => {
                selectAndClose(keyResult.id);
              }}
            >
              <Flex align="center" className="min-w-0 gap-2">
                <OKRIcon className="h-[1.1rem] shrink-0" strokeWidth={2.4} />
                <Text className="truncate">{keyResult.name}</Text>
              </Flex>
              {keyResult.id === keyResultId ? (
                <CheckIcon className="h-4 w-4 shrink-0" />
              ) : null}
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

export const ObjectiveKeyResultSubMenu = ({
  disabled = false,
  keyResultId,
  objectiveId,
  onChange,
  teamId,
}: {
  disabled?: boolean;
  keyResultId: string | null;
  objectiveId: string | null;
  onChange: (selection: ObjectiveKeyResultSelection) => void;
  teamId: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data: objectives = [], isPending } = useTeamObjectives(
    teamId,
    deferredQuery,
  );

  return (
    <Menu.SubMenu
      onOpenChange={(open) => {
        if (!open) setQuery("");
      }}
    >
      <Menu.SubTrigger className="justify-between gap-4" disabled={disabled}>
        <Flex align="center" className="gap-2">
          <ObjectiveIcon className="h-[1.1rem]" />
          <Text>
            {getTermDisplay("objectiveTerm", { capitalize: true })} and{" "}
            {getTermDisplay("keyResultTerm")}
          </Text>
        </Flex>
        <ArrowRightIcon
          className="text-text-muted h-3.5 w-3.5"
          strokeWidth={2.4}
        />
      </Menu.SubTrigger>
      <Menu.SubItems className="w-80 max-w-[calc(100vw-2rem)]">
        <Box className="px-3 pt-0.5 pb-1.5">
          <Menu.Input
            autoFocus
            onChange={(event) => {
              setQuery(event.target.value);
            }}
            placeholder={`Find ${getTermDisplay("objectiveTerm")}...`}
            value={query}
          />
        </Box>
        <Divider className="my-1.5" />
        <Menu.Group>
          <Menu.Item
            active={!objectiveId}
            className="justify-between gap-3"
            onSelect={() => {
              onChange({ objectiveId: null, keyResultId: null });
            }}
          >
            <Flex align="center" className="gap-2">
              <ObjectiveIcon className="h-[1.1rem]" />
              <Text>No {getTermDisplay("objectiveTerm")}</Text>
            </Flex>
            {!objectiveId ? <CheckIcon className="h-4 w-4" /> : null}
          </Menu.Item>
        </Menu.Group>
        <Menu.Separator />
        <Menu.Group className="max-h-80 overflow-y-auto">
          {isPending ? <MenuLoadingSkeleton rows={5} /> : null}
          {!isPending && objectives.length === 0 ? (
            <Text className="px-2 py-2" color="muted">
              No {getTermDisplay("objectiveTerm")} found.
            </Text>
          ) : null}
          {objectives.map((objective) => (
            <ObjectiveMenuItem
              currentKeyResultId={keyResultId}
              currentObjectiveId={objectiveId}
              key={objective.id}
              objective={objective}
              onSelect={onChange}
            />
          ))}
        </Menu.Group>
      </Menu.SubItems>
    </Menu.SubMenu>
  );
};
