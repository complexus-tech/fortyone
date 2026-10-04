import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { WipLimitIndicator } from "./wip-limit-indicator";

describe("WIP limit indicator", () => {
  it.each([undefined, null, 0])("hides an unset limit (%s)", (limit) => {
    const { container } = render(
      <WipLimitIndicator activeCount={6} limit={limit} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it.each([
    { count: 4, description: "4 tasks, limit 5" },
    { count: 5, description: "5 tasks, limit 5, at capacity" },
  ])(
    "keeps $count/5 advisory without an over-limit warning",
    ({ count, description }) => {
      render(<WipLimitIndicator activeCount={count} limit={5} />);
      const indicator = screen.getByText(`${count}/5`).parentElement;
      expect(screen.getByText(description)).toBeInTheDocument();
      expect(indicator).not.toHaveClass("text-warning");
      expect(indicator?.querySelector("svg")).toBeNull();
    },
  );

  it("shows warning text and an icon only after exceeding the limit", () => {
    render(<WipLimitIndicator activeCount={6} limit={5} />);
    const indicator = screen.getByText("6/5").parentElement;
    expect(indicator).toHaveClass("text-warning");
    expect(indicator?.querySelector("svg")).toHaveClass("text-warning");
    expect(indicator?.querySelector("svg")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
    expect(
      screen.getByText("6 tasks, limit 5, over limit by 1"),
    ).toBeInTheDocument();
  });

  it("uses clear status copy and the configured plural term in its tooltip", async () => {
    render(
      <WipLimitIndicator
        activeCount={9}
        limit={5}
        pluralTaskTerm="tickets"
        taskTerm="tickets"
      />,
    );
    fireEvent.pointerMove(screen.getByText("9/5").parentElement!, {
      pointerType: "mouse",
    });
    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip).toHaveTextContent(
      "9 tickets in this status · limit 5. Reduce by 4 to get within the limit.",
    );
    expect(tooltip).toHaveTextContent("Includes tickets hidden by filters.");
  });

  it("opens its description on keyboard focus and closes on Escape while preserving focus", async () => {
    render(<WipLimitIndicator activeCount={6} limit={5} />);
    const indicator = screen.getByText("6/5").parentElement!;
    expect(indicator).toHaveAttribute("tabindex", "0");

    act(() => {
      indicator.focus();
    });
    expect(indicator).toHaveFocus();
    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip).toHaveTextContent(
      "6 tasks in this status · limit 5. Reduce by 1 to get within the limit.",
    );
    expect(tooltip).toHaveTextContent("Includes tasks hidden by filters.");
    expect(indicator).toHaveAttribute("aria-describedby", tooltip.id);
    expect(
      screen.getByText("6 tasks, limit 5, over limit by 1"),
    ).toBeInTheDocument();

    fireEvent.keyDown(indicator, { key: "Escape" });
    await waitFor(() => {
      expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();
    });
    expect(indicator).toHaveFocus();
  });
});
