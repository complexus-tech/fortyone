import { fireEvent, render, screen } from "@testing-library/react";
import { DesktopDraftNotice } from "./desktop-draft-notice";

const actions = {
  onRestore: jest.fn(),
  onDiscard: jest.fn(),
  onRetry: jest.fn(),
};

beforeEach(() => jest.clearAllMocks());

it.each(["Saving draft…", "Draft saved on this device"])(
  "keeps routine draft status quiet: %s",
  (status) => {
    const view = render(
      <DesktopDraftNotice
        {...actions}
        error={null}
        recovered={false}
        status={status}
      />,
    );
    expect(view.container).toBeEmptyDOMElement();
  },
);

it("keeps recovered work available for an explicit restore or discard", () => {
  render(<DesktopDraftNotice {...actions} error={null} recovered status="" />);
  fireEvent.click(screen.getByRole("button", { name: "Restore draft" }));
  expect(actions.onRestore).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole("button", { name: "Discard saved draft" }));
  expect(actions.onDiscard).toHaveBeenCalledTimes(1);
});

it("keeps failed writes visible with a retry action", () => {
  render(
    <DesktopDraftNotice
      {...actions}
      error="Not enough disk space"
      recovered={false}
      status="Draft not saved"
    />,
  );
  expect(screen.getByRole("alert")).toHaveTextContent("Not enough disk space");
  fireEvent.click(screen.getByRole("button", { name: "Retry saving" }));
  expect(actions.onRetry).toHaveBeenCalledTimes(1);
});
