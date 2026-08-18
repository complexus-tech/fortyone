import * as React from "react";
import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Box, Container } from "ui";
import { SidebarToggleButton } from "./sidebar/sidebar-toggle-button";

const injectSidebarToggle = (children: React.ReactNode) => {
  const childArray = React.Children.toArray(children);
  const firstChild = childArray[0];

  if (!React.isValidElement<{ children?: React.ReactNode }>(firstChild)) {
    return children;
  }

  return [
    React.cloneElement(
      firstChild,
      { key: firstChild.key ?? "header-content" },
      <>
        <Box className="hidden shrink-0 md:flex">
          <SidebarToggleButton />
        </Box>
        {firstChild.props.children}
      </>,
    ),
    ...childArray.slice(1),
  ];
};

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "stick border-border top-0 z-10 flex h-[3.6rem] w-full items-center border-b-[0.5px]",
        className,
      )}
    >
      {injectSidebarToggle(children)}
    </Container>
  );
};
