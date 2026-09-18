import { render, screen } from "@testing-library/react";
import { CalendarEventDescription } from "./calendar-event-description";

it("renders calendar HTML as formatted content with safe links", () => {
  const { container } = render(
    <CalendarEventDescription description='<h3>Agenda</h3><p>Review <strong>results</strong><br>Then plan.</p><ul><li>Next steps</li></ul><a href="https://example.com/notes">Meeting notes</a>' />,
  );
  expect(screen.getByRole("heading", { name: "Agenda" })).toBeInTheDocument();
  expect(container.querySelector("strong")).toHaveTextContent("results");
  expect(container.querySelector("br")).toBeInTheDocument();
  expect(screen.getByRole("listitem")).toHaveTextContent("Next steps");
  expect(screen.getByRole("link", { name: "Meeting notes" })).toHaveAttribute(
    "target",
    "_blank",
  );
  expect(screen.getByRole("link", { name: "Meeting notes" })).toHaveAttribute(
    "rel",
    "noopener noreferrer",
  );
});

it("removes scripts, embedded content, handlers and unsafe URLs from provider HTML", () => {
  const { container } = render(
    <CalendarEventDescription description='<div onclick="alert(1)" style="position:fixed">Agenda</div><script>alert(1)</script><iframe src="https://example.com"></iframe><img src="https://example.com/tracker" onerror="alert(1)"><a href="javascript:alert(1)">Unsafe</a>' />,
  );
  expect(screen.getByText("Agenda")).toBeInTheDocument();
  expect(
    container.querySelector("script,iframe,img,[onclick],[onerror],[style]"),
  ).toBeNull();
  expect(screen.getByText("Unsafe")).not.toHaveAttribute("href");
});

it("preserves literal plain text, email brackets and line breaks", () => {
  const description = "Contact <person@example.com>\nBudget < 500 & questions";
  const { container } = render(
    <CalendarEventDescription description={description} />,
  );
  expect(container.firstChild).toHaveTextContent(
    "Contact <person@example.com>",
  );
  expect(container.firstChild?.textContent).toBe(description);
  expect(container.firstChild).toHaveClass("whitespace-pre-wrap");
});
