"use client";
import { BreadCrumbs, Flex } from "ui";
import { SprintsIcon } from "icons";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useTerminology } from "@/hooks";

export const RunningSprintsHeader = () => {
  const { getTermDisplay } = useTerminology();

  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: getTermDisplay("sprintTerm", {
                variant: "plural",
                capitalize: true,
              }),
              icon: <SprintsIcon className="h-[1.1rem] w-auto" />,
            },
          ]}
          className="md:hidden"
        />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: `Current ${getTermDisplay("sprintTerm", {
                variant: "plural",
                capitalize: true,
              })}`,
              icon: <SprintsIcon className="h-[1.1rem] w-auto" />,
            },
          ]}
          className="hidden md:flex"
        />
      </Flex>
    </HeaderContainer>
  );
};
