/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { Link } from "@/types";
import { Links } from "./links";

jest.mock("@/hooks", () => ({
  useUserRole: () => ({ userRole: "member" }),
}));

jest.mock("@/hooks/clipboard", () => ({
  useCopyToClipboard: () => [false, jest.fn()],
}));

jest.mock("@/lib/hooks/delete-link-mutation", () => ({
  useDeleteLinkMutation: () => ({ mutateAsync: jest.fn() }),
}));

jest.mock("@/lib/hooks/figma", () => ({
  useStoryFigmaLinks: () => ({ data: [] }),
}));

jest.mock("@/lib/hooks/link-metadata", () => ({
  useLinkMetadata: () => ({ data: undefined }),
}));

jest.mock("@/modules/google-drive/public/files", () => ({
  useGoogleDriveFiles: () => ({ data: [] }),
}));

jest.mock("./add-link-dialog", () => ({
  AddLinkDialog: () => null,
}));

const googleSheetLink: Link = {
  createdAt: "2026-09-04T08:00:00Z",
  id: "link-1",
  storyId: "story-1",
  title: "Launch plan",
  updatedAt: "2026-09-04T08:00:00Z",
  url: "https://docs.google.com/spreadsheets/d/google-sheet-1/edit",
};

describe("Links Google Drive presentation", () => {
  it("uses the native icon without an Add preview step or redundant type label", () => {
    render(
      <Links
        isLinksOpen
        links={[googleSheetLink]}
        setIsLinksOpen={jest.fn()}
        storyId="story-1"
      />,
    );

    expect(screen.getByText("Launch plan")).toBeInTheDocument();
    expect(screen.getByText("Google Drive link")).toBeInTheDocument();
    expect(screen.queryByText("Google Sheet")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Add preview" }),
    ).not.toBeInTheDocument();
    expect(
      document.querySelector('svg[viewBox="0 0 192 192"] path[fill="#009954"]'),
    ).not.toBeNull();
  });
});
