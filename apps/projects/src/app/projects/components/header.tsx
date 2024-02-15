import { BreadCrumbs, Button, Flex } from "ui";
import { Plus, Search, Settings2 } from "lucide-react";
import { useState } from "react";
import { HeaderContainer } from "@/components/layout";
import { NewProjectDialog } from "@/components/ui";

export const Header = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Projects",
            url: "/projects",
          },
        ]}
      />
      <Flex gap={3}>
        <Button
          align="center"
          className="px-[0.6rem]"
          color="tertiary"
          leftIcon={<Search className="h-[1.1rem] w-auto" />}
          size="sm"
          variant="outline"
        >
          <span className="sr-only">Search</span>
        </Button>
        <Button
          color="tertiary"
          leftIcon={<Settings2 className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <Button
          leftIcon={<Plus className="h-5 w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
        >
          New project
        </Button>
      </Flex>
      <NewProjectDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
