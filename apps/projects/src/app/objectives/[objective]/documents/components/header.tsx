import { BreadCrumbs, Button, Flex } from "ui";
import { useState } from "react";
import {
  DocsIcon,
  PlusIcon,
  PreferencesIcon,
  ObjectiveIcon,
  SearchIcon,
} from "icons";
import { HeaderContainer } from "@/components/shared";
import { NewObjectiveDialog } from "@/components/ui";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Web Design",
            icon: <ObjectiveIcon className="h-4 w-auto" />,
          },
          {
            name: "Docs",
            icon: <DocsIcon className="h-4 w-auto" />,
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
        <Button
          leftIcon={<PlusIcon className="h-5 w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
        >
          New wiki
        </Button>
      </Flex>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
