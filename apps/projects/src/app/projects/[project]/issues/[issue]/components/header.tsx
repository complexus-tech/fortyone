import { Bell, ChevronUp, ChevronDown, Star } from "lucide-react";
import { BreadCrumbs, Button, Flex, Text } from "ui";
import { HeaderContainer } from "@/components/layout";

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
            <ChevronUp className="h-4 w-auto" />
          </Button>
          <Button className="mr-2" color="tertiary" disabled size="sm">
            <ChevronDown className="h-4 w-auto" />
          </Button>
          <Button
            className="px-3"
            color="tertiary"
            leftIcon={<Star className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Favourite
          </Button>
          <Button
            className="px-2"
            leftIcon={<Bell className="h-4 w-auto" />}
            size="sm"
          >
            Subscribe
          </Button>
        </Flex>
      </Flex>
    </HeaderContainer>
  );
};
