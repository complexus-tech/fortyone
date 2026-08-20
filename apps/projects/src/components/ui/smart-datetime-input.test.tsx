/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { parseSmartDateTime, SmartDateTimeInput } from "./smart-datetime-input";

const referenceDate = new Date(2026, 7, 20, 10, 0, 0);

describe("parseSmartDateTime", () => {
  it("resolves natural-language dates in the local timezone", () => {
    const result = parseSmartDateTime("tomorrow at 3pm", referenceDate);

    expect(result.error).toBeNull();
    expect(result.date).toEqual(new Date(2026, 7, 21, 15, 0, 0));
  });

  it("requires an explicit time", () => {
    expect(parseSmartDateTime("tomorrow", referenceDate)).toEqual({
      date: null,
      error: "missing-time",
    });
  });
});

describe("SmartDateTimeInput", () => {
  it("uses the shared input while accepting natural-language dates", async () => {
    const onChange = jest.fn();
    render(
      <SmartDateTimeInput
        label="Start"
        onChange={onChange}
        referenceDate={referenceDate}
        value="2026-08-20T11:00"
      />,
    );

    const input = screen.getByDisplayValue("Aug 20, 2026 at 11:00 AM");
    fireEvent.change(input, { target: { value: "tomorrow at 3pm" } });
    fireEvent.blur(input);

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith("2026-08-21T15:00");
    });
  });

  it("keeps the native datetime picker as a synchronized fallback", () => {
    const onChange = jest.fn();
    render(
      <SmartDateTimeInput
        label="End"
        onChange={onChange}
        value="2026-08-20T12:00"
      />,
    );

    fireEvent.change(screen.getByLabelText("Choose end"), {
      target: { value: "2026-08-20T13:30" },
    });

    expect(onChange).toHaveBeenCalledWith("2026-08-20T13:30");
    expect(
      screen.getByDisplayValue("Aug 20, 2026 at 1:30 PM"),
    ).toBeInTheDocument();
  });

  it("reports incomplete natural-language values before submission", () => {
    const onValidityChange = jest.fn();
    render(
      <SmartDateTimeInput
        label="Start"
        onChange={jest.fn()}
        onValidityChange={onValidityChange}
        referenceDate={referenceDate}
        value="2026-08-20T11:00"
      />,
    );

    const input = screen.getByDisplayValue("Aug 20, 2026 at 11:00 AM");
    fireEvent.change(input, { target: { value: "tomorrow" } });
    expect(onValidityChange).toHaveBeenLastCalledWith(false);

    fireEvent.blur(input);
    expect(
      screen.getByText("Include a time, for example “tomorrow at 3pm”."),
    ).toBeInTheDocument();
  });
});
