import { BreadCrumbs, Button, Flex } from "ui";
import { IssuesIcon, PreferencesIcon } from "icons";
import { HeaderContainer } from "@/components/layout";
import { NewIssueButton, SideDetailsSwitch } from "@/components/ui";

export const Header = ({
  isExpanded,
  setIsExpanded,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
}) => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "My issues",
            icon: <IssuesIcon className="h-5 w-auto" />,
          },
          { name: "Assigned" },
        ]}
      />
      <Flex gap={2}>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <NewIssueButton />
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
