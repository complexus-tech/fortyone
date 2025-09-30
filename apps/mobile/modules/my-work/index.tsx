import React from "react";
import { Header } from "./components/header";
import { StoriesList } from "./components/list";
import { SafeContainer } from "@/components/ui";

export const MyWork = () => {
  return (
    <SafeContainer isFull>
      <Header />
      <StoriesList />
    </SafeContainer>
  );
};
