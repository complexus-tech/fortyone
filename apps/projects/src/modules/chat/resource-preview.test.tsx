import type { ReactNode } from "react";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import {
  NativeLinkNavigationProvider,
  useNativeLinkNavigation,
} from "@/components/ui/link-preview/native-navigation";
import type { ChatResource } from "./types";
import { ResourcePreview } from "./resource-preview";

jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Dialog: Object.assign(Container, {
      Content: Container,
      Header: Container,
      Title: Container,
      Description: Container,
      Body: Container,
      Footer: Container,
    }),
    Button: ({
      children,
      onClick,
    }: {
      children: ReactNode;
      onClick: () => void;
    }) => (
      <button onClick={onClick} type="button">
        {children}
      </button>
    ),
  };
});
jest.mock("./link-preview", () => ({
  ChatLinkPreview: () => {
    const navigate = useNativeLinkNavigation();
    return (
      <button
        onClick={(event) => {
          navigate?.(
            {
              kind: "story",
              id: "TM-276",
              workspaceSlug: "acme",
              href: "https://acme.fortyone.app/work/TM-276",
            },
            event.currentTarget,
          );
        }}
        type="button"
      >
        View story
      </button>
    );
  },
}));
jest.mock("./resource-presentation", () => ({
  ResourceIcon: () => null,
  resourceMetadata: () => "Story",
  resourceTitle: () => "Story link",
}));

const resource = {
  kind: "link",
  id: "link",
  messageId: "message",
  url: "https://acme.fortyone.app/work/TM-276",
  title: "Story link",
  createdAt: "2026-10-17T09:00:00Z",
  author: { fullName: "Alex" },
} as ChatResource;
afterEach(cleanup);

it.each([true, false])(
  "closes the link dialog only when navigation is handled: %s",
  (handled) => {
    const onClose = jest.fn();
    const navigate = jest.fn(() => handled);
    const returningToMessage = { current: false };
    render(
      <NativeLinkNavigationProvider onOpen={navigate}>
        <ResourcePreview
          conversationName="Product"
          onClose={onClose}
          onViewMessage={jest.fn()}
          resource={resource}
          returningToMessage={returningToMessage}
          triggerRef={{ current: null }}
          url={resource.url}
        />
      </NativeLinkNavigationProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "View story" }));
    expect(navigate).toHaveBeenCalledTimes(1);
    expect(onClose).toHaveBeenCalledTimes(handled ? 1 : 0);
    expect(returningToMessage.current).toBe(handled);
  },
);
