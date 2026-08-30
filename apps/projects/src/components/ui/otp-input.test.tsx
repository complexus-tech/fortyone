/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { OTPInput } from "./otp-input";

describe("OTPInput", () => {
  it("normalizes multi-character input to its final digit and advances focus", () => {
    const onChange = jest.fn();

    render(<OTPInput length={4} onChange={onChange} value="" />);

    const firstInput = screen.getByRole("textbox", {
      name: "Verification code digit 1 of 4",
    });
    const secondInput = screen.getByRole("textbox", {
      name: "Verification code digit 2 of 4",
    });
    fireEvent.change(firstInput, { target: { value: "12" } });

    expect(onChange).toHaveBeenCalledWith("2");
    expect(secondInput).toHaveFocus();
  });

  it("does not emit a change for non-numeric input", () => {
    const onChange = jest.fn();

    render(<OTPInput onChange={onChange} value="" />);

    fireEvent.change(screen.getByRole("textbox", { name: /digit 1 of 6/i }), {
      target: { value: "x" },
    });

    expect(onChange).not.toHaveBeenCalled();
  });
});
