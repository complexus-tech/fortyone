/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { ActivityActor } from "./activity-actor";

jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children, href }: { children: ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

jest.mock("ui", () => {
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );

  return {
    Avatar: ({ name }: { name: string }) => <span>{name}</span>,
    Box: Container,
    Button: ({ children, href }: { children: ReactNode; href?: string }) =>
      href ? (
        <a href={href}>{children}</a>
      ) : (
        <button type="button">{children}</button>
      ),
    Flex: Container,
    Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
    Tooltip: ({
      children,
      title,
    }: {
      children: ReactNode;
      title: ReactNode;
    }) => (
      <>
        {children}
        {title}
      </>
    ),
  };
});

jest.mock("./maya-avatar", () => ({
  MayaAvatar: ({ name }: { name: string }) => <span>{name}</span>,
}));

describe("ActivityActor", () => {
  it("keeps a member profile route scoped to the active workspace", () => {
    render(
      <ActivityActor
        displayName="Taylor Reed"
        displayUsername="taylor"
        isSelfActivity={false}
        member={{
          avatarUrl: null,
          fullName: "Taylor Reed",
          id: "member-1",
          username: "taylor",
        }}
        withWorkspace={(path) => `/acme${path}`}
      />,
    );

    expect(screen.getByRole("link", { name: "Go to profile" })).toHaveAttribute(
      "href",
      "/acme/profile/member-1",
    );
  });

  it("keeps system actors non-navigable and labels Maya as the AI agent", () => {
    render(
      <ActivityActor
        displayName="Maya"
        displayUsername="maya"
        isSelfActivity={false}
        member={{
          avatarUrl: null,
          fullName: "Maya",
          id: "maya",
          isSystem: true,
          username: "maya",
        }}
        withWorkspace={(path) => `/acme${path}`}
      />,
    );

    expect(screen.getByText("(AI Agent)")).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Go to profile" }),
    ).not.toBeInTheDocument();
  });
});
