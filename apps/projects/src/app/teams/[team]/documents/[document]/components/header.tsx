import { Avatar, BreadCrumbs, Button, Flex, Text } from "ui";
import { DocsIcon, StarIcon } from "icons";
import { HeaderContainer } from "@/components/shared";

export const Header = () => {
  return (
    <HeaderContainer>
      <Flex align="center" className="w-full" justify="between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Engineering",
              icon: "🚀",
            },
            {
              name: "Documents",
              icon: <DocsIcon className="h-[1.1rem] w-auto" />,
            },
            {
              name: "Sprint Planning",
            },
          ]}
        />
        <Flex align="center" gap={3} justify="between">
          <Text as="span" color="muted">
            Saved
          </Text>
          <Avatar
            name="Joseph Mukorivo"
            size="sm"
            src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
          />
          <Button
            className="aspect-square"
            color="tertiary"
            leftIcon={<StarIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            <span className="sr-only">Favorite</span>
          </Button>
          <Button className="px-4" size="sm">
            Publish
          </Button>
          <Button className="px-3" color="tertiary" size="sm" variant="outline">
            Save draft
          </Button>
        </Flex>
      </Flex>
    </HeaderContainer>
  );
};
