import { BreadCrumbs, Button, Flex } from "ui";
import { DocsIcon, PlusIcon, PreferencesIcon, SearchIcon } from "icons";
import { HeaderContainer } from "@/components/shared";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Engineering",
            icon: "🚀",
          },
          {
            name: "Documents",
            icon: <DocsIcon className="h-5 w-auto" />,
          },
        ]}
      />
      <Flex gap={3}>
        <Button
          align="center"
          className="px-[0.6rem]"
          color="tertiary"
          leftIcon={<SearchIcon className="h-[1.1rem] w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="sm">
          New Document
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
