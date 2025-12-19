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
import { toast } from "sonner";
import { RowWrapper } from "@/components/ui";
import { useRemoveAssociationMutation } from "@/modules/story/hooks/remove-association-mutation";
import type { StoryAssociation } from "../types";
import { useTeams } from "@/modules/teams/hooks/teams";
import Link from "next/link";
import { slugify } from "@/utils";

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
      className="px-1 text-sm font-bold uppercase dark:border-dark-100 dark:text-opacity-70"
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

  return (
    <Box className="mt-8">
      {associations.length > 0 && (
        <Flex
          align="center"
          className={cn({
            "border-b-[0.5px] border-gray-100/60 pb-2 dark:border-dark-200":
              !isAssociationsOpen,
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
        <Box className="mt-2 border-t-[0.5px] border-gray-100 pb-0 dark:border-dark-100">
          {associations.map((assoc) => {
            const teamCode = teams.find(
              (team) => team.id === assoc.story.teamId,
            )?.code;

            return (
              <RowWrapper className="gap-8 px-1 py-2 md:px-1" key={assoc.id}>
                <Flex align="center" className="flex-1 gap-2">
                  <AssociationBadge association={assoc} storyId={storyId} />
                  <Link
                    href={`/story/${assoc.story.id}/${slugify(assoc.story.title)}`}
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
                        <Menu.Item active onSelect={() => {}}>
                          <LinkIcon /> Related...
                        </Menu.Item>
                        <Menu.Item onSelect={() => {}}>
                          <WarningIcon />
                          Blocks...
                        </Menu.Item>
                        <Menu.Item onSelect={() => {}}>
                          <ErrorIcon />
                          Blocked by...
                        </Menu.Item>
                        <Menu.Item onSelect={() => {}}>
                          <DuplicateIcon /> Duplicate of...
                        </Menu.Item>
                        <Menu.Item onSelect={() => {}}>
                          <CopyIcon /> Duplicated by...
                        </Menu.Item>
                      </Menu.Group>
                      <Menu.Separator />
                      <Menu.Group>
                        <Menu.Item
                          className="text-error tracking-wide"
                          onSelect={() => {
                            removeAssociation({
                              associationId: assoc.id,
                              storyId,
                            }).then(() => {
                              toast.success("Success", {
                                description: "Association removed",
                              });
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
