import { BreadCrumbs, Button, Flex, Text } from "ui";
import { HeaderContainer } from "@/components/layout";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  BellIcon,
  StarIcon,
} from "@/components/icons";

export const Header = () => {
  return (
    <HeaderContainer>
      <Flex align="center" className="w-full" justify="between">
        <BreadCrumbs
          breadCrumbs={[
            { name: "Complexus" },
            { name: "Web design" },
            {
              name: "COM-12",
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
          <Button color="tertiary" size="sm">
            <ArrowUpIcon className="h-4 w-auto" />
          </Button>
          <Button className="mr-2" color="tertiary" disabled size="sm">
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
