import { BreadCrumbs, Button, Flex } from "ui";
import { PlusIcon, PreferencesIcon, ArrowDownIcon, EpicsIcon } from "icons";
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
            name: "Epics",
            icon: <EpicsIcon className="h-[1.15rem] w-auto" />,
          },
        ]}
      />
      <Flex gap={3}>
        <Button
          color="tertiary"
          leftIcon={<PreferencesIcon className="h-4 w-auto" />}
          rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
          size="sm"
          variant="outline"
        >
          Display
        </Button>
        <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="sm">
          New Epic
        </Button>
      </Flex>
    </HeaderContainer>
  );
};
