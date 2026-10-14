// The camera, nodes and connections are real app components. These cards make
// the isolated fixture deterministic without auth, menus, or production data.
const Card = ({
  title,
  width,
  height,
  kind = "objectives",
}: {
  title: string;
  width: number;
  height: number;
  kind?: string;
}) => (
  <div className="fixture-card" style={{ width, height }}>
    <div className="fixture-eyebrow">{kind}</div>
    <h3>{title}</h3>
    <p>Measure strategy execution and improve customer outcomes at scale.</p>
    <div className="fixture-properties">
      <span>In progress</span>
      <span>High</span>
      <span>Oct 14</span>
    </div>
    <div className="fixture-progress">
      <span />
    </div>
    <div className="fixture-properties">
      <span>Product</span>
      <span>2 key results</span>
    </div>
  </div>
);

export const ObjectiveNodeCard = ({
  objective,
}: {
  objective: { id: string; name: string };
}) => (
  <Card
    title={objective.name}
    width={340}
    height={218 + (Number(objective.id.split("-")[1]) % 3) * 19}
  />
);
export const KeyResultNodeCard = ({
  keyResult,
}: {
  keyResult: { name: string };
}) => (
  <Card title={keyResult.name} width={280} height={154} kind="key result" />
);
export const PillarNodeCard = ({ name }: { name: string }) => (
  <Card title={name} width={340} height={168} kind="pillar" />
);
export const UltimateGoalNodeCard = ({ title }: { title: string }) => (
  <Card title={title} width={440} height={236} kind="ultimate goal" />
);
