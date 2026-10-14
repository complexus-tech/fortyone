/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { useParams } from "next/navigation";
import { useUserRole } from "@/hooks";
import {
  useCreateKeyResultMutation,
  useKeyResults,
  useObjective,
} from "../../hooks";
import { useOkrQualityAssessment } from "../../hooks/use-okr-quality-assessment";
import { NewKeyResultButton, NewKeyResultDialog } from "./new-key-result";

jest.mock("next/navigation", () => ({ useParams: jest.fn() }));
jest.mock("@/hooks", () => ({
  useMediaQuery: () => false,
  useTerminology: () => ({
    getTermDisplay: (_term: string, options?: { capitalize?: boolean }) =>
      options?.capitalize ? "Key result" : "key result",
  }),
  useUserRole: jest.fn(),
}));
jest.mock("../../hooks", () => ({
  useCreateKeyResultMutation: jest.fn(),
  useKeyResults: jest.fn(),
  useObjective: jest.fn(),
}));
jest.mock("../../hooks/use-okr-quality-assessment", () => ({
  useOkrQualityAssessment: jest.fn(),
}));
jest.mock("../../components/okr-quality-banner", () => ({
  OkrQualityBanner: () => null,
}));

const mutate = jest.fn();

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useParams).mockReturnValue({ objectiveId: "route-objective" });
  jest.mocked(useUserRole).mockReturnValue({ userRole: "admin" });
  jest.mocked(useCreateKeyResultMutation).mockReturnValue({
    mutate,
  } as unknown as ReturnType<typeof useCreateKeyResultMutation>);
  jest.mocked(useObjective).mockReturnValue({
    data: undefined,
  } as ReturnType<typeof useObjective>);
  jest.mocked(useKeyResults).mockReturnValue({
    data: [],
  } as unknown as ReturnType<typeof useKeyResults>);
  jest.mocked(useOkrQualityAssessment).mockReturnValue({
    assessment: null,
    isAssessing: false,
  } as ReturnType<typeof useOkrQualityAssessment>);
});

describe("key result creation dialog", () => {
  it("defers data and quality hooks until the controlled dialog opens", () => {
    const onOpenChange = jest.fn();
    const { rerender } = render(
      <NewKeyResultDialog
        isOpen={false}
        objectiveId="map-objective"
        onOpenChange={onOpenChange}
      />,
    );

    expect(useObjective).not.toHaveBeenCalled();
    expect(useKeyResults).not.toHaveBeenCalled();
    expect(useOkrQualityAssessment).not.toHaveBeenCalled();
    expect(useCreateKeyResultMutation).not.toHaveBeenCalled();

    rerender(
      <NewKeyResultDialog
        isOpen
        objectiveId="map-objective"
        onOpenChange={onOpenChange}
      />,
    );

    expect(screen.getByRole("dialog")).toBeVisible();
    expect(useObjective).toHaveBeenCalledWith("map-objective");
    expect(useKeyResults).toHaveBeenCalledWith("map-objective", true);
    expect(useOkrQualityAssessment).toHaveBeenCalled();
  });

  it("uses the selected map objective when submitting and preserves numeric starting progress", () => {
    const onOpenChange = jest.fn();
    render(
      <NewKeyResultDialog
        isOpen
        objectiveId="map-objective"
        onOpenChange={onOpenChange}
      />,
    );

    fireEvent.change(screen.getByLabelText(/Name$/), {
      target: { value: "Increase weekly activation" },
    });
    fireEvent.change(screen.getByLabelText(/Starting Value$/), {
      target: { value: "12" },
    });
    fireEvent.change(screen.getByLabelText(/Target Value$/), {
      target: { value: "20" },
    });
    fireEvent.change(screen.getByLabelText(/Start Date$/), {
      target: { value: "2026-10-14" },
    });
    fireEvent.change(screen.getByLabelText(/Deadline$/), {
      target: { value: "2026-11-14" },
    });
    const form = screen.getByLabelText(/Name$/).closest("form");
    expect(form).not.toBeNull();
    fireEvent.submit(form!);

    expect(mutate).toHaveBeenCalledWith(
      expect.objectContaining({
        objectiveId: "map-objective",
        name: "Increase weekly activation",
        measurementType: "number",
        startValue: 12,
        targetValue: 20,
        currentValue: 12,
      }),
    );
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it.each([
    ["guest", "map-objective"],
    ["admin", ""],
  ] as const)(
    "does not mount creation for role %s and objective %s",
    (userRole, objectiveId) => {
      jest.mocked(useUserRole).mockReturnValue({ userRole });
      render(
        <NewKeyResultDialog
          isOpen
          objectiveId={objectiveId}
          onOpenChange={jest.fn()}
        />,
      );

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      expect(useObjective).not.toHaveBeenCalled();
      expect(useCreateKeyResultMutation).not.toHaveBeenCalled();
      expect(mutate).not.toHaveBeenCalled();
    },
  );

  it("resets the draft when the selected objective changes", () => {
    const onOpenChange = jest.fn();
    const { rerender } = render(
      <NewKeyResultDialog
        isOpen
        objectiveId="first-objective"
        onOpenChange={onOpenChange}
      />,
    );
    fireEvent.change(screen.getByLabelText(/Name$/), {
      target: { value: "First objective draft" },
    });

    rerender(
      <NewKeyResultDialog
        isOpen
        objectiveId="second-objective"
        onOpenChange={onOpenChange}
      />,
    );

    expect(screen.getByLabelText(/Name$/)).toHaveValue("");
    expect(useObjective).toHaveBeenLastCalledWith("second-objective");
  });
});

describe("existing key result creation button", () => {
  it.each([
    [undefined, "route-objective"],
    ["explicit-objective", "explicit-objective"],
  ])(
    "resolves objective %s to %s only after opening",
    (objectiveId, expectedId) => {
      render(<NewKeyResultButton objectiveId={objectiveId} />);
      expect(useObjective).not.toHaveBeenCalled();

      fireEvent.click(screen.getByRole("button", { name: "Add Key result" }));

      expect(useObjective).toHaveBeenCalledWith(expectedId);
      expect(useKeyResults).toHaveBeenCalledWith(expectedId, true);
    },
  );

  it("keeps cancellation resetting the form for the next opening", () => {
    render(<NewKeyResultButton />);
    fireEvent.click(screen.getByRole("button", { name: "Add Key result" }));
    fireEvent.change(screen.getByLabelText(/Name$/), {
      target: { value: "Discarded draft" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Add Key result" }));
    expect(screen.getByLabelText(/Name$/)).toHaveValue("");
    expect(mutate).not.toHaveBeenCalled();
  });
});
