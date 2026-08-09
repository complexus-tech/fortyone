import { render, screen } from "@testing-library/react";
import { SlackIcon } from "icons";
import { RequestIntegrationBanner } from "./details";

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
