import React from "react";
import { Flex } from "ui";
import { ComplexusLogo } from "@/components/ui/logo";

export const AiIcon = () => {
  return (
    <Flex
      align="center"
      className="size-8 rounded-full bg-dark dark:bg-white"
      justify="center"
    >
      <ComplexusLogo className="h-4 text-white dark:text-dark" />
    </Flex>
  );
};
