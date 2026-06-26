"use client";
import { Flex, Button, Text, Box, Menu, Badge } from "ui";
import {
  ArrowDown2Icon,
  ArrowUp2Icon,
  CopyIcon,
  DeleteIcon,
  DuplicateIcon,
  ErrorIcon,
  LinkIcon,
  MoreHorizontalIcon,
  WarningIcon,
} from "icons";
import { cn } from "lib";
import Link from "next/link";
import type { ReactNode } from "react";
import { RowWrapper } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import { useRemoveAssociationMutation } from "@/modules/story/hooks/remove-association-mutation";
import { useUpdateAssociationMutation } from "@/modules/story/hooks/update-association-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { slugify } from "@/utils";
import type { StoryAssociation, StoryAssociationType } from "../types";

type AssociationOption = {
  direction: "incoming" | "outgoing";
  icon: ReactNode;
  label: string;
  type: StoryAssociationType;
};

const ASSOCIATION_OPTIONS: AssociationOption[] = [
  {
    direction: "outgoing",
    icon: <LinkIcon />,
    label: "Related...",
    type: "related",
  },
  {
    direction: "outgoing",
    icon: <WarningIcon />,
    label: "Blocks...",
    type: "blocking",
  },
  {
    direction: "incoming",
    icon: <ErrorIcon />,
    label: "Blocked by...",
    type: "blocking",
  },
  {
    direction: "outgoing",
    icon: <DuplicateIcon />,
    label: "Duplicate of...",
    type: "duplicate",
  },
  {
    direction: "incoming",
    icon: <CopyIcon />,
    label: "Duplicated by...",
    type: "duplicate",
  },
];

const AssociationBadge = ({
  association,
  storyId,
}: {
  association: StoryAssociation;
  storyId: string;
}) => {
  const isOutgoing = association.fromStoryId === storyId;

  const getDetails = (): {
    label: string;
    color: "warning" | "danger" | "tertiary";
  } => {
    switch (association.type) {
      case "blocking":
        return {
          label: isOutgoing ? "Blocks" : "Blocked by",
          color: isOutgoing ? "warning" : "danger",
        };
      case "duplicate":
        return {
          label: isOutgoing ? "Duplicate of" : "Duplicated by",
          color: "tertiary",
        };
      default:
        return {
          label: "Related to",
          color: "tertiary",
        };
    }
  };

  const { label, color } = getDetails();

  return (
    <Badge
      className="d dark:text-opacity-70 shrink-0 px-2 text-[0.8125rem] font-bold uppercase"
      color={color}
      rounded="sm"
    >
      {label}
    </Badge>
  );
};

export const Associations = ({
  storyId,
  associations,
  isAssociationsOpen,
  setIsAssociationsOpen,
}: {
  storyId: string;
  associations: StoryAssociation[];
  isAssociationsOpen: boolean;
  setIsAssociationsOpen: (isOpen: boolean) => void;
}) => {
  const { data: teams = [] } = useTeams();
  const { mutateAsync: removeAssociation } = useRemoveAssociationMutation();
  const { mutate: updateAssociation } = useUpdateAssociationMutation();
  const { withWorkspace } = useWorkspacePath();

  return (
    <Box className="mt-4">
      {associations.length > 0 && (
        <Flex
          align="center"
          className={cn({
            "border-border d border-b-[0.5px] pb-2": !isAssociationsOpen,
          })}
          justify="between"
        >
          <Button
            className="font-semibold"
            color="tertiary"
            leftIcon={<LinkIcon className="mr-0.5 h-5" />}
            onClick={() => {
              setIsAssociationsOpen(!isAssociationsOpen);
            }}
            rightIcon={
              isAssociationsOpen ? (
                <ArrowDown2Icon className="h-4" />
              ) : (
                <ArrowUp2Icon className="h-4" />
              )
            }
            size="sm"
            variant="naked"
          >
            Associations ({associations.length})
          </Button>
        </Flex>
      )}

      {isAssociationsOpen && associations.length > 0 ? (
        <Box className="border-border d mt-2 border-t-[0.5px] pb-0">
          {associations.map((assoc) => {
            const teamCode = teams.find(
              (team) => team.id === assoc.story.teamId,
            )?.code;
            const associatedStoryId =
              assoc.fromStoryId === storyId
                ? assoc.toStoryId
                : assoc.fromStoryId;
            const getAssociationPayload = (option: AssociationOption) =>
              option.direction === "incoming"
                ? {
                    associationId: assoc.id,
                    fromStoryId: associatedStoryId,
                    storyId,
                    toStoryId: storyId,
                    type: option.type,
                  }
                : {
                    associationId: assoc.id,
                    fromStoryId: storyId,
                    storyId,
                    toStoryId: associatedStoryId,
                    type: option.type,
                  };
            const isOptionActive = (option: AssociationOption) =>
              assoc.type === option.type &&
              (option.type === "related" ||
                (option.direction === "outgoing"
                  ? assoc.fromStoryId === storyId
                  : assoc.toStoryId === storyId));

            return (
              <RowWrapper className="gap-8 px-1 py-2 md:px-1" key={assoc.id}>
                <Flex align="center" className="flex-1 gap-2">
                  <AssociationBadge association={assoc} storyId={storyId} />
                  <Link
                    href={withWorkspace(
                      `/story/${assoc.story.id}/${slugify(assoc.story.title)}`,
                    )}
                  >
                    <Text
                      className="line-clamp-1 font-medium"
                      title={assoc.story.title}
                    >
                      <span className="mr-2 opacity-60">
                        {teamCode}-{assoc.story.sequenceId}
                      </span>
                      {assoc.story.title}
                    </Text>
                  </Link>
                </Flex>
                <Flex align="center" className="shrink-0" gap={3}>
                  <Menu>
                    <Menu.Button>
                      <Button
                        asIcon
                        color="tertiary"
                        rounded="full"
                        size="sm"
                        variant="naked"
                      >
                        <MoreHorizontalIcon />
                      </Button>
                    </Menu.Button>
                    <Menu.Items align="end" className="min-w-72">
                      <Menu.Group className="px-3 py-1">
                        <Text
                          color="muted"
                          fontSize="sm"
                          fontWeight="bold"
                          transform="uppercase"
                        >
                          Change association type
                        </Text>
                      </Menu.Group>
                      <Menu.Separator />
                      <Menu.Group>
                        {ASSOCIATION_OPTIONS.map((option) => (
                          <Menu.Item
                            active={isOptionActive(option)}
                            key={`${option.direction}-${option.type}`}
                            onSelect={() => {
                              updateAssociation(getAssociationPayload(option));
                            }}
                          >
                            {option.icon}
                            {option.label}
                          </Menu.Item>
                        ))}
                      </Menu.Group>
                      <Menu.Separator />
                      <Menu.Group>
                        <Menu.Item
                          className="text-error tracking-wide"
                          onSelect={() => {
                            removeAssociation({
                              associationId: assoc.id,
                              storyId,
                            });
                          }}
                        >
                          <DeleteIcon />
                          Remove association...
                        </Menu.Item>
                      </Menu.Group>
                    </Menu.Items>
                  </Menu>
                </Flex>
              </RowWrapper>
            );
          })}
        </Box>
      ) : null}
    </Box>
  );
};
