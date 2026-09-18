import type * as ReactTypes from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { DocumentIndex } from "./document-index";

jest.mock("icons", () => ({ ListIcon: () => null }));
jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactTypes>("react");
  const PassThrough = ({ children }: { children: React.ReactNode }) => children;
  return {
    Button: ({ children }: { children: React.ReactNode }) =>
      React.createElement("button", { type: "button" }, children),
    Text: PassThrough,
    Popover: Object.assign(PassThrough, {
      Trigger: PassThrough,
      Content: PassThrough,
    }),
  };
});

beforeEach(() => {
  global.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
  window.matchMedia = jest.fn().mockReturnValue({ matches: true });
});

function fixture() {
  const container = document.createElement("div");
  container.innerHTML =
    "<h2>Outside the document</h2><article><h2>Overview</h2><h3>Details</h3><h2> </h2></article>";
  document.body.append(container);
  const scrollTo = jest.fn();
  container.scrollTo = scrollTo;
  return { container, scrollTo };
}

test("tracks edited, inserted and removed headings without including surrounding chrome", async () => {
  const { container } = fixture();
  const view = render(
    <DocumentIndex contentSelector="article" scrollContainer={container} />,
  );
  await waitFor(() => {
    expect(screen.getAllByRole("button", { name: "Details" })).toHaveLength(2);
  });
  expect(
    screen.queryByRole("button", { name: "Outside the document" }),
  ).not.toBeInTheDocument();
  container.querySelector("h3")!.textContent = "Next steps";
  await waitFor(() => {
    expect(screen.getAllByRole("button", { name: "Next steps" })).toHaveLength(
      2,
    );
  });
  expect(
    screen.queryByRole("button", { name: "Details" }),
  ).not.toBeInTheDocument();
  container.querySelector("article")!.innerHTML = "<h2>Only one section</h2>";
  await waitFor(() => {
    expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
  });
  view.unmount();
  container.remove();
});

test("navigates duplicate heading names by element within the correct scroll container", async () => {
  const { container, scrollTo } = fixture();
  container.querySelector("article")!.innerHTML =
    "<h2>Notes</h2><h2>Notes</h2>";
  const secondHeading = container.querySelectorAll("article h2")[1];
  secondHeading.getBoundingClientRect = () => ({ top: 600 }) as DOMRect;
  container.getBoundingClientRect = () => ({ top: 100 }) as DOMRect;
  const view = render(
    <DocumentIndex
      contentSelector="article"
      readingOffset={112}
      scrollContainer={container}
    />,
  );
  await waitFor(() => {
    expect(screen.getAllByRole("button", { name: "Notes" })).toHaveLength(4);
  });
  fireEvent.click(screen.getAllByRole("button", { name: "Notes" })[1]);
  expect(scrollTo).toHaveBeenCalledWith({
    top: 388,
    behavior: "instant",
  });
  view.unmount();
  container.remove();
});
