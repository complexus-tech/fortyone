"use client";
import { Badge, BreadCrumbs, Button, Flex, Text } from "ui";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  BellIcon,
  StarIcon,
  StoryIcon,
  UndoIcon,
} from "icons";
import { HeaderContainer } from "@/components/shared";
import { useRestoreStoryMutation } from "@/modules/story/hooks/restore-mutation";
import { useTeams } from "@/lib/hooks/teams";

export const Header = ({
  sequenceId,
  teamId,
  isDeleted,
  storyId,
}: {
  sequenceId: number;
  teamId: string;
  isDeleted: boolean;
  storyId: string;
}) => {
  const { data: teams = [] } = useTeams();
  const { name, code } = teams.find((team) => team.id === teamId)!!;
  const { mutateAsync } = useRestoreStoryMutation();

  const restoreStory = async () => {
    mutateAsync(storyId);
  };

  return (
    <HeaderContainer>
      <Flex align="center" className="w-full" justify="between">
        <Flex align="center" gap={3}>
          <BreadCrumbs
            breadCrumbs={[
              {
                name,
                icon: "🚀",
                url: "/teams/web",
              },
              {
                name: "Stories",
                icon: <StoryIcon className="h-[1.1rem] w-auto" />,
              },
              {
                name: `${code?.toUpperCase()}-${sequenceId}`,
              },
            ]}
          />
          {isDeleted && (
            <Badge className="uppercase" color="tertiary" rounded="full">
              Deleted
            </Badge>
          )}
        </Flex>
        <Flex align="center" gap={2} justify="between">
          {isDeleted && (
            <Button
              size="sm"
              color="tertiary"
              onClick={restoreStory}
              leftIcon={<UndoIcon className="h-4 w-auto" />}
            >
              Restore story
            </Button>
          )}
          <Text className="mr-2">
            2 /{" "}
            <Text as="span" color="muted">
              8
            </Text>
          </Text>
          <Button
            className="aspect-square"
            color="tertiary"
            rounded="xl"
            disabled={isDeleted}
            size="sm"
          >
            <ArrowUpIcon className="h-4 w-auto" />
          </Button>
          <Button
            className="mr-10 aspect-square"
            color="tertiary"
            disabled
            rounded="xl"
            size="sm"
          >
            <ArrowDownIcon className="h-4 w-auto" />
          </Button>
          <Button
            className="aspect-square"
            color="tertiary"
            disabled={isDeleted}
            size="sm"
          >
            <StarIcon className="h-4 w-auto" />
            <span className="sr-only">Favourite</span>
          </Button>
          <Button
            className="aspect-square"
            color="tertiary"
            disabled={isDeleted}
            size="sm"
          >
            <BellIcon className="h-5 w-auto" />
            <span className="sr-only">Subscribe</span>
          </Button>
        </Flex>
      </Flex>
    </HeaderContainer>
  );
};
