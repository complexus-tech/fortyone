import { act, fireEvent, render, screen } from "@testing-library/react";
import { useObjectiveDragNavigation } from "./use-objective-drag-navigation";

const Harness = ({ enabled = true }: { enabled?: boolean }) => {
  const navigation = useObjectiveDragNavigation(enabled);
  return (
    <>
      <button
        onClick={() => {
          navigation.start("first");
        }}
        type="button"
      >
        Start drag
      </button>
      <button onClick={navigation.release} type="button">
        Release drag
      </button>
      <a data-objective-link="first" href="#first">
        <span>First objective</span>
      </a>
      <a data-objective-link="second" href="#second">
        Second objective
      </a>
    </>
  );
};

describe("objective anchor navigation after dragging", () => {
  beforeEach(() => jest.useFakeTimers());
  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });
  it("prevents the anchor default even when Dnd-kit stops clicks at document capture", () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Start drag" }));
    fireEvent.click(screen.getByRole("button", { name: "Release drag" }));
    const stopAtDocument = (event: MouseEvent) => {
      event.stopPropagation();
    };
    document.addEventListener("click", stopAtDocument, true);
    const event = new MouseEvent("click", { bubbles: true, cancelable: true });
    screen.getByText("First objective").dispatchEvent(event);
    document.removeEventListener("click", stopAtDocument, true);
    expect(event.defaultPrevented).toBe(true);
  });

  it("blocks only the dragged objective and expires the release guard", () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Start drag" }));
    fireEvent.click(screen.getByRole("button", { name: "Release drag" }));
    const other = new MouseEvent("click", { bubbles: true, cancelable: true });
    screen.getByRole("link", { name: "Second objective" }).dispatchEvent(other);
    expect(other.defaultPrevented).toBe(false);
    act(() => {
      jest.advanceTimersByTime(51);
    });
    const afterRelease = new MouseEvent("click", {
      bubbles: true,
      cancelable: true,
    });
    screen
      .getByRole("link", { name: "First objective" })
      .dispatchEvent(afterRelease);
    expect(afterRelease.defaultPrevented).toBe(false);
  });

  it("removes its capture guard when leaving the board", () => {
    const { rerender } = render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: "Start drag" }));
    rerender(<Harness enabled={false} />);
    const event = new MouseEvent("click", { bubbles: true, cancelable: true });
    screen.getByRole("link", { name: "First objective" }).dispatchEvent(event);
    expect(event.defaultPrevented).toBe(false);
  });
});
