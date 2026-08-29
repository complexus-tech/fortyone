/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import { SlackIcon } from "icons";
import { RequestIntegrationBanner } from "./details";

jest.mock("react-markdown", () => ({
  __esModule: true,
  default: ({ children }: { children: string }) => <>{children}</>,
}));
jest.mock("remark-gfm", () => ({ __esModule: true, default: jest.fn() }));

describe("RequestIntegrationBanner", () => {
  it("keeps intake actions available when the provider has no source URL", () => {
    render(
      <RequestIntegrationBanner
        canEditRequest
        icon={<SlackIcon />}
        onAccept={jest.fn()}
        onDecline={jest.fn()}
        openLabel="Open on Slack"
        primaryText="Request from Slack"
        secondaryText="#product"
      />,
    );

    expect(screen.getByLabelText("Intake actions")).toBeInTheDocument();
    expect(screen.queryByTitle("Open on Slack")).not.toBeInTheDocument();
  });
});
