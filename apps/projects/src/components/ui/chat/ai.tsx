import React from "react";
import { Flex } from "ui";
import { ComplexusLogo } from "@/components/ui/logo";

export const AiIcon = () => {
  return (
    <Flex
      align="center"
      className="size-8 rounded-full bg-primary bg-gradient-to-r from-primary to-secondary/50"
      justify="center"
    >
      <ComplexusLogo className="h-4 text-white" />
    </Flex>
  );
};
