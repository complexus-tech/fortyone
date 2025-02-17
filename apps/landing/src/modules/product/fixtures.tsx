import { SprintsIcon, ObjectiveIcon, StoryIcon, OKRIcon } from "icons";
import storyCard from "../../../public/kanban.png";

export const features = [
  {
    icon: (
      <StoryIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Stories",
    title: "Manage and Track Tasks",
    overview:
      "Transform tasks into stories to manage and track your project with precision. Capture every detail and streamline your workflow for optimal team collaboration.",
    breakdown: [
      {
        title: "Story Properties",
        overview:
          "Easily assign priorities, labels, estimates, relations, references, and due dates to each story, ensuring comprehensive task management.",
      },
      {
        title: "Rapid Story Management",
        overview:
          "Quickly create, edit, and update stories, enhancing efficiency and reducing time spent on task administration.",
      },
      {
        title: "Parent and Sub-stories",
        overview:
          "Break down large stories into smaller, manageable pieces or group related stories, improving task organization and clarity.",
      },
      {
        title: "Story Templates",
        overview:
          "Utilize predefined templates for faster story creation and consistent information sharing across the team.",
      },
      {
        title: "Duplicate Detection",
        overview:
          "Automatically identify and merge similar stories using AI, maintaining a clean and organized workspace.",
      },
      {
        title: "Story Importer",
        overview:
          "Effortlessly import stories from other issue trackers, ensuring a smooth transition and continuity in project management.",
      },
    ],
  },
  {
    icon: (
      <ObjectiveIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Objectives",
    title: "Set and Achieve Goals",
    overview:
      "Define clear objectives to guide your project. Monitor progress and ensure your team stays on the right path.",
    breakdown: [
      {
        title: "Clear Goal Setting",
        overview:
          "Define specific, measurable, and achievable goals to provide clear direction for your project using Objective Key Results (OKRs).",
      },
      {
        title: "Progress Monitoring",
        overview:
          "Track the progress and health of objectives to ensure alignment with overall project goals and timely completion.",
      },
    ],
  },
  {
    icon: (
      <OKRIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "OKRs",
    title: "Align and Drive",
    overview:
      "Align your team and drive towards a common goal with OKRs. Monitor progress and ensure your team stays on the right path.",
    breakdown: [
      {
        title: "Align and Drive",
        overview:
          "Align your team and drive towards a common goal with OKRs. Monitor progress and ensure your team stays on the right path.",
      },
    ],
  },
  {
    icon: (
      <SprintsIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Sprints",
    title: "Iterate and Deliver",
    overview:
      "Organize your work into sprints for focused, iterative development and consistent delivery.",
    breakdown: [
      {
        title: "Task Segmentation",
        overview:
          "Break down stories into manageable sprints, promoting continuous improvement and agile methodologies.",
      },
      {
        title: "Sprint Delivery Management",
        overview:
          "Manage the delivery of features and functionalities within each sprint, ensuring efficient workflow and timely completion.",
      },
    ],
  },
];
