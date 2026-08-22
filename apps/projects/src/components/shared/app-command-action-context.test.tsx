/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import {
  AppCommandActionProvider,
  useAppCommandAction,
  useCurrentAppCommandAction,
} from "./app-command-action-context";

const RegisteredAction = ({ onSelect }: { onSelect: () => void }) => {
  useAppCommandAction({
    id: "test:create",
    label: "Create item",
    onSelect,
  });

  return null;
};

const CurrentAction = () => {
  const action = useCurrentAppCommandAction();

  return action ? (
    <button onClick={action.onSelect} type="button">
      {action.label}
    </button>
  ) : (
    <span>No action</span>
  );
};

describe("AppCommandActionProvider", () => {
  it("registers, invokes, and clears a contextual command action", () => {
    const onSelect = jest.fn();
    const { rerender } = render(
      <AppCommandActionProvider>
        <RegisteredAction onSelect={onSelect} />
        <CurrentAction />
      </AppCommandActionProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Create item" }));
    expect(onSelect).toHaveBeenCalledTimes(1);

    rerender(
      <AppCommandActionProvider>
        <CurrentAction />
      </AppCommandActionProvider>,
    );

    expect(screen.getByText("No action")).toBeInTheDocument();
  });
});
