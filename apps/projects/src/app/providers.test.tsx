/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { QueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { render } from "@testing-library/react";
import { getQueryClient } from "./get-query-client";
import { Providers } from "./providers";

const mockObservedClients: unknown[] = [];

jest.mock("@tanstack/react-query", () => ({
  QueryClientProvider: ({
    children,
    client,
  }: {
    children: ReactNode;
    client: unknown;
  }) => {
    mockObservedClients.push(client);
    return <>{children}</>;
  },
}));

jest.mock("./get-query-client", () => ({
  getQueryClient: jest.fn(),
}));

jest.mock("./posthog", () => ({
  PostHogProvider: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

jest.mock("./posthog-page-view", () => () => null);

jest.mock("next-themes", () => ({
  ThemeProvider: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

jest.mock("nuqs/adapters/next/app", () => ({
  NuqsAdapter: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

const mockGetQueryClient = jest.mocked(getQueryClient);
const mockQueryClient = {} as QueryClient;

describe("Providers", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockObservedClients.splice(0);
    mockGetQueryClient.mockReturnValue(mockQueryClient);
  });

  it("keeps one query client across provider rerenders", () => {
    const { rerender } = render(
      <Providers>
        <div>First render</div>
      </Providers>,
    );

    rerender(
      <Providers>
        <div>Second render</div>
      </Providers>,
    );

    expect(mockGetQueryClient).toHaveBeenCalledTimes(1);
    expect(mockObservedClients).toEqual([mockQueryClient, mockQueryClient]);
  });
});
