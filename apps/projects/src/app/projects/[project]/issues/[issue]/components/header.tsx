import { BreadCrumbs, Button, Flex, Text } from "ui";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  BellIcon,
  StarIcon,
  ProjectsIcon,
  IssuesIcon,
} from "icons";
import { HeaderContainer } from "@/components/layout";

export const Header = () => {
  return (
    <HeaderContainer>
      <Flex align="center" className="w-full" justify="between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "All projects",
              icon: <ProjectsIcon className="h-4 w-auto" />,
              url: "/projects",
            },
            {
              name: "Web design",
              icon: "🚀",
              url: "/projects/web",
            },
            {
              name: "Issues",
              icon: <IssuesIcon className="h-[1.1rem] w-auto" />,
            },
            {
              name: "Web-12",
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
            className="mr-2 aspect-square"
            color="tertiary"
            disabled
            rounded="xl"
            size="sm"
          >
            <ArrowDownIcon className="h-4 w-auto" />
          </Button>
          <Button
            className="px-3"
            color="tertiary"
            leftIcon={<StarIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Favourite
          </Button>
          <Button
            className="px-2"
            leftIcon={<BellIcon className="h-[1.15rem] w-auto" />}
            size="sm"
          >
            Subscribe
          </Button>
        </Flex>
      </Flex>
    </HeaderContainer>
  );
};
