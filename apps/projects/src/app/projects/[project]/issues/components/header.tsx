import { BreadCrumbs, Button, Flex } from "ui";
import { Columns3, Settings2 } from "lucide-react";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton } from "@/components/ui";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Web design",
          },
          { name: "Assigned" },
        ]}
      />
      <Flex gap={2}>
        <Button
          color="tertiary"
          leftIcon={<Columns3 className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Board
        </Button>
        <Button
          color="tertiary"
          leftIcon={<Settings2 className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <NewIssueButton />
      </Flex>
    </HeaderContainer>
  );
};
