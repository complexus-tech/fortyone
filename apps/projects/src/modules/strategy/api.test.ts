import { normalizeStrategyMap } from "./normalize-strategy-map";

describe("normalizeStrategyMap", () => {
  it("normalizes nullable strategy collections", () => {
    expect(
      normalizeStrategyMap({
        ultimateGoal: "Grow responsibly",
        description: null,
        pillars: [
          {
            id: "pillar-1",
            name: "Customer trust",
            description: null,
            orderIndex: 0,
            objectiveIds: null,
          },
        ],
      }),
    ).toEqual({
      ultimateGoal: "Grow responsibly",
      description: null,
      pillars: [
        {
          id: "pillar-1",
          name: "Customer trust",
          description: null,
          orderIndex: 0,
          objectiveIds: [],
        },
      ],
    });
  });

  it("returns an empty strategy when the response has no data", () => {
    expect(normalizeStrategyMap(null)).toEqual({
      ultimateGoal: "",
      description: null,
      pillars: [],
    });
  });
});
