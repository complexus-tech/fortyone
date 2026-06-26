import { render, screen } from "@testing-library/react";
import { TeamOptionsProvider, useTeamOptions } from "./provider";

const mockUseLocalStorage = jest.fn(
  <T,>(key: string, initialValue: T): [T, (value: T) => void] => [
    initialValue,
    jest.fn(),
  ],
);

jest.mock("@/hooks", () => ({
  useLocalStorage: (...args: Parameters<typeof mockUseLocalStorage>) =>
    mockUseLocalStorage(...args),
}));

jest.mock("@/components/ui/stories-filter-state", () => ({
  useStoriesFilters: () => ({
    filters: {},
    resetFilters: jest.fn(),
    setFilters: jest.fn(),
  }),
}));

const BacklogOptionsProbe = () => {
  const { viewOptions } = useTeamOptions();

  return (
    <output data-testid="show-sub-stories">
      {String(viewOptions.showSubStories)}
    </output>
  );
};

describe("Team backlog options", () => {
  beforeEach(() => {
    mockUseLocalStorage.mockClear();
    localStorage.clear();
  });

  it("shows backlog sub-stories by default with a migrated storage key", () => {
    render(
      <TeamOptionsProvider layout="list">
        <BacklogOptionsProbe />
      </TeamOptionsProvider>,
    );

    expect(screen.getByTestId("show-sub-stories")).toHaveTextContent("true");
    expect(mockUseLocalStorage).toHaveBeenCalledWith(
      "teams:backlog:view-options:v2:list",
      expect.objectContaining({ showSubStories: true }),
    );
  });
});
