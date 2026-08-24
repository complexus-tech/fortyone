import calendarWeekDark from "../../../public/images/product/calendar-week-dark.webp";
import calendarWeekLight from "../../../public/images/product/calendar-week-light.webp";
import mayaHomeDark from "../../../public/images/product/maya-home-dark.webp";
import mayaHomeLight from "../../../public/images/product/maya-home-light.webp";
import mayaObjectiveRisksDark from "../../../public/images/product/maya-objective-risks-dark.webp";
import mayaObjectiveRisksLight from "../../../public/images/product/maya-objective-risks-light.webp";
import { FeatureProductWorkflow } from "./feature-product-workflow";

const PLANNING_WORKFLOW = [
  {
    alt: "Maya ready to answer a planning question using the work already in FortyOne",
    darkImage: mayaHomeDark,
    description:
      "Ask in plain language. Maya starts from the goals, tasks, workload, and connected project context already in FortyOne.",
    label: "Ask in context",
    lightImage: mayaHomeLight,
    title: "Start with the decision the team needs to make.",
    value: "ask-in-context",
    url: "https://fortyone.app/maya",
  },
  {
    alt: "Maya identifying at-risk objectives and the delivery decisions that need attention",
    darkImage: mayaObjectiveRisksDark,
    description:
      "Bring capacity, timing, dependencies, and goal health into view before new work is added to the plan.",
    label: "See the tradeoffs",
    lightImage: mayaObjectiveRisksLight,
    title: "Understand what the next move will affect.",
    value: "see-the-tradeoffs",
    url: "https://fortyone.app/summary",
  },
  {
    alt: "FortyOne weekly calendar showing project work planned around the team's existing commitments",
    darkImage: calendarWeekDark,
    description:
      "Use workload and calendar availability to find a realistic start window before the team commits to more work.",
    label: "Find a work window",
    lightImage: calendarWeekLight,
    title: "Place the work where the team can deliver it.",
    value: "find-a-work-window",
    url: "https://fortyone.app/calendar",
  },
] as const;

export function AiPlanningWorkflow() {
  return (
    <FeatureProductWorkflow
      ariaLabel="Explore the AI planning workflow"
      description="Maya turns the context already around your work into a planning recommendation your team can understand and control."
      heading="Follow the path from question to approved plan."
      id="ai-planning-workflow"
      items={PLANNING_WORKFLOW}
    />
  );
}
