/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
} from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { TimeNeededMenu } from "./time-needed-menu";

jest.mock("ui", () => {
  const Container = ({ children }: { children?: ReactNode }) => <>{children}</>;
  const Button = ({
    active: _active,
    children,
    leftIcon,
    rightIcon,
    type = "button",
    ...props
  }: ButtonHTMLAttributes<HTMLButtonElement> & {
    active?: boolean;
    leftIcon?: ReactNode;
    rightIcon?: ReactNode;
  }) => (
    <button {...props} type={type === "submit" ? "submit" : "button"}>
      {leftIcon}
      {children}
      {rightIcon}
    </button>
  );
  const Input = (props: InputHTMLAttributes<HTMLInputElement>) => (
    <input {...props} />
  );
  const Popover = ({ children }: { children?: ReactNode }) => <>{children}</>;
  Popover.Trigger = Container;
  Popover.Content = Container;

  return {
    Box: Container,
    Button,
    Divider: () => <hr />,
    Flex: Container,
    Input,
    Popover,
    Text: Container,
  };
});

describe("TimeNeededMenu", () => {
  it("selects a preset as schedulable minutes", () => {
    const setTimeNeeded = jest.fn();
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items setTimeNeeded={setTimeNeeded} />
      </TimeNeededMenu>,
    );

    fireEvent.click(screen.getByRole("button", { name: "2h" }));

    expect(setTimeNeeded).toHaveBeenCalledWith({
      estimatedDurationMinutes: 120,
      minimumFocusBlockMinutes: null,
    });
  });

  it("uses hours as the default custom input unit", () => {
    const setTimeNeeded = jest.fn();
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items setTimeNeeded={setTimeNeeded} />
      </TimeNeededMenu>,
    );

    expect(screen.getByRole("button", { name: "hrs" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
    fireEvent.change(screen.getByLabelText("Custom time needed"), {
      target: { value: "1.5" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Set" }));

    expect(setTimeNeeded).toHaveBeenCalledWith({
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: null,
    });
  });

  it("lets users schedule automatically around open time", () => {
    const setTimeNeeded = jest.fn();
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items
          estimatedDurationMinutes={120}
          minimumFocusBlockMinutes={60}
          setTimeNeeded={setTimeNeeded}
        />
      </TimeNeededMenu>,
    );

    expect(document.body).toHaveTextContent(
      "Automatic fills open time. Or choose a minimum block size.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Automatic" }));

    expect(setTimeNeeded).toHaveBeenCalledWith({
      estimatedDurationMinutes: 120,
      minimumFocusBlockMinutes: null,
    });
  });

  it("offers a two-hour focus block", () => {
    const setTimeNeeded = jest.fn();
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items
          estimatedDurationMinutes={8 * 60}
          setTimeNeeded={setTimeNeeded}
        />
      </TimeNeededMenu>,
    );

    const twoHourButtons = screen.getAllByRole("button", { name: "2h" });
    const twoHourFocusBlock = twoHourButtons.at(-1);
    expect(twoHourFocusBlock).toBeDefined();
    fireEvent.click(twoHourFocusBlock as HTMLButtonElement);

    expect(setTimeNeeded).toHaveBeenCalledWith({
      estimatedDurationMinutes: 8 * 60,
      minimumFocusBlockMinutes: 120,
    });
  });

  it("keeps an explicit short focus block visible after duration is reduced", () => {
    const setTimeNeeded = jest.fn();
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items
          estimatedDurationMinutes={30}
          minimumFocusBlockMinutes={15}
          setTimeNeeded={setTimeNeeded}
        />
      </TimeNeededMenu>,
    );

    expect(document.body).toHaveTextContent("Focus blocks");
    fireEvent.click(screen.getByRole("button", { name: "Automatic" }));

    expect(setTimeNeeded).toHaveBeenCalledWith({
      estimatedDurationMinutes: 30,
      minimumFocusBlockMinutes: null,
    });
  });

  it("shows an explicit non-preset focus block", () => {
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items
          estimatedDurationMinutes={60}
          minimumFocusBlockMinutes={20}
          setTimeNeeded={jest.fn()}
        />
      </TimeNeededMenu>,
    );

    expect(screen.getByRole("button", { name: "20m" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );
  });

  it("explains the maximum when custom time is outside the supported range", () => {
    const setTimeNeeded = jest.fn();
    render(
      <TimeNeededMenu>
        <TimeNeededMenu.Trigger>
          <button type="button">Open</button>
        </TimeNeededMenu.Trigger>
        <TimeNeededMenu.Items setTimeNeeded={setTimeNeeded} />
      </TimeNeededMenu>,
    );

    fireEvent.change(screen.getByLabelText("Custom time needed"), {
      target: { value: "41" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Set" }));

    expect(setTimeNeeded).not.toHaveBeenCalled();
    expect(document.body).toHaveTextContent(
      "Enter a time between 1 minute and 40 hours.",
    );
  });
});
