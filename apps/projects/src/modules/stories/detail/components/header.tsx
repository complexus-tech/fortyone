import { BreadCrumbs, Button, Flex, Text } from "ui";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  BellIcon,
  StarIcon,
  StoryIcon,
} from "icons";
import { HeaderContainer } from "@/components/shared";

export const Header = ({ sequenceId }: { sequenceId: number }) => {
  return (
    <HeaderContainer>
      <Flex align="center" className="w-full" justify="between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Engineering",
              icon: "🚀",
              url: "/teams/web",
            },
            {
              name: "Stories",
              icon: <StoryIcon className="h-[1.1rem] w-auto" />,
            },
            {
              name: `Web-${sequenceId}`,
            },
          ]}
        />
        <Flex align="center" gap={2} justify="between">
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
          <Button className="aspect-square" color="tertiary" size="sm">
            <StarIcon className="h-4 w-auto" />
            <span className="sr-only">Favourite</span>
          </Button>
          <Button className="aspect-square" color="tertiary" size="sm">
            <BellIcon className="h-5 w-auto" />
            <span className="sr-only">Subscribe</span>
          </Button>
        </Flex>
      </Flex>
    </HeaderContainer>
  );
};
