import { BreadCrumbs, Button, Flex } from "ui";
import { Columns3, SlidersHorizontal } from "lucide-react";
import { NewIssueButton } from "@/components/ui";
import { HeaderContainer } from "@/components/layout";

export const Header = () => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "Dashboard",
            icon: <Columns3 className="h-5 w-auto" />,
          },
        ]}
      />
      <Flex gap={2}>
        <Button
          color="tertiary"
          leftIcon={<SlidersHorizontal className="h-4 w-auto" />}
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
