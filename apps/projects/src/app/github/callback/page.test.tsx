/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import GitHubCallbackPage from "./page";

const mockUseSearchParams = jest.fn();

jest.mock("next/navigation", () => ({
  useSearchParams: () => mockUseSearchParams(),
}));

jest.mock("@/components/ui", () => ({
  Logo: () => <div aria-hidden="true" />,
}));

jest.mock("ui", () => ({
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
}));

describe("GitHubCallbackPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("keeps the existing callback loading state visible while search params suspend", () => {
    const pendingSearchParams = new Promise<never>((_resolve) => {
      // Intentionally remains pending so the Suspense fallback stays visible.
    });
    const suspendedSearchParams = Object.assign(
      new Error("Search params are pending"),
      {
        then: pendingSearchParams.then.bind(pendingSearchParams),
      },
    );
    mockUseSearchParams.mockImplementation(() => {
      throw suspendedSearchParams;
    });

    render(<GitHubCallbackPage />);

    expect(
      screen.getByText("Connecting your GitHub account..."),
    ).toBeInTheDocument();
  });
});
