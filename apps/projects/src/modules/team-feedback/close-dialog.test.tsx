/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { CloseTeamFeedbackDialog } from "./close-dialog";

jest.mock("ui", () => {
  const React = jest.requireActual("react");
  const Container = ({ children, ...props }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const Dialog = ({ children }: { children: ReactNode }) =>
    React.createElement("div", null, children);
  Dialog.Content = Container;
  Dialog.Header = Container;
  Dialog.Body = Container;
  Dialog.Footer = Container;
  const DialogTitle = ({ children }: { children: ReactNode }) =>
    React.createElement("h2", null, children);
  Dialog.Title = DialogTitle;

  const Button = ({
    children,
    color: _color,
    loading: _loading,
    loadingText: _loadingText,
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    color?: string;
    loading?: boolean;
    loadingText?: string;
  }) => React.createElement("button", { type: "button", ...props }, children);
  const Text = ({
    children,
    color: _color,
    ...props
  }: ComponentPropsWithoutRef<"span"> & { color?: string }) =>
    React.createElement("span", props, children);
  const TextArea = ({
    label,
    ...props
  }: ComponentPropsWithoutRef<"textarea"> & { label?: string }) =>
    React.createElement(
      "label",
      null,
      label,
      React.createElement("textarea", props),
    );

  return { Button, Dialog, Flex: Container, Text, TextArea };
});

describe("CloseTeamFeedbackDialog", () => {
  it("submits a trimmed public explanation", () => {
    const onConfirm = jest.fn();

    render(
      <CloseTeamFeedbackDialog
        isLoading={false}
        onCancel={jest.fn()}
        onConfirm={onConfirm}
      />,
    );

    fireEvent.change(screen.getByLabelText("Public explanation (optional)"), {
      target: {
        value: "  This request conflicts with the current workflow.  ",
      },
    });
    fireEvent.click(screen.getByRole("button", { name: "Close feedback" }));

    expect(onConfirm).toHaveBeenCalledWith(
      "This request conflicts with the current workflow.",
    );
  });

  it("uses null when feedback is closed quietly", () => {
    const onConfirm = jest.fn();

    render(
      <CloseTeamFeedbackDialog
        isLoading={false}
        onCancel={jest.fn()}
        onConfirm={onConfirm}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Close feedback" }));

    expect(onConfirm).toHaveBeenCalledWith(null);
  });
});
