export type StrategicPillar = {
  id: string;
  name: string;
  description: string | null;
  orderIndex: number;
  objectiveIds: string[];
};

export type StrategyMap = {
  ultimateGoal: string;
  description: string | null;
  pillars: StrategicPillar[];
};

export type StrategyUpdate = Pick<StrategyMap, "ultimateGoal" | "description">;

export type NewStrategicPillar = Pick<
  StrategicPillar,
  "name" | "description" | "orderIndex"
>;

export type UpdateStrategicPillar = Partial<NewStrategicPillar>;
