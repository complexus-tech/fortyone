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
