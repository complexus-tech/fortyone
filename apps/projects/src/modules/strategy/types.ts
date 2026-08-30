import type { StrategicPillar, StrategyMap } from "@/shared/strategy-map/types";

export type { StrategicPillar, StrategyMap };

export type StrategyMember = {
  avatarUrl: string | null;
  fullName: string;
  id: string;
  username: string;
};

export type StrategyUpdate = Pick<StrategyMap, "ultimateGoal" | "description">;

export type NewStrategicPillar = Pick<
  StrategicPillar,
  "name" | "description" | "orderIndex"
>;

export type UpdateStrategicPillar = Partial<NewStrategicPillar>;
