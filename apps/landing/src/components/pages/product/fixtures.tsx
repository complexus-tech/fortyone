import {
  AnalyticsIcon,
  ChatIcon,
  DocsIcon,
  EpicsIcon,
  SprintsIcon,
  ObjectiveIcon,
  RoadmapIcon,
  StoryIcon,
  WhiteboardIcon,
} from "icons";
import storyCard from "../../../../public/kanban.png";

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
      <RoadmapIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Roadmaps",
    title: "Plan Your Objectives' Journey",
    overview:
      "Craft and visualize detailed roadmaps to plan and track your objectives' journey from start to finish.",
    breakdown: [
      {
        title: "Visual Timeline Planning",
        overview:
          "Create visual representations of objectives, sprints and timelines to facilitate effective planning and resource allocation.",
      },
      {
        title: "Milestone Tracking",
        overview:
          "Monitor the progress of sprints and overall milestones, ensuring timely delivery and adherence to plans.",
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
  {
    icon: (
      <EpicsIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Epics",
    title: "Oversee Major Initiatives",
    overview:
      "Oversee and track the progress of large-scale features or initiatives, ensuring all critical components stay on track.",
    breakdown: [
      {
        title: "Large-scale Feature Management",
        overview:
          "Oversee the development and progress of significant features, ensuring all aspects are covered.",
      },
      {
        title: "Initiative Alignment",
        overview:
          "Track the progress and alignment of major initiatives, helping to achieve strategic objectives.",
      },
    ],
  },
  {
    icon: (
      <DocsIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Documents",
    title: "Store and Collaborate",
    overview:
      "Store, organize, and collaborate on project documents seamlessly. Keep all your important files in one place.",
    breakdown: [
      {
        title: "Centralized Document Storage",
        overview:
          "Safely store and organize all project-related documents for easy access and retrieval.",
      },
      {
        title: "Real-time Collaboration",
        overview:
          "Collaborate in real-time with team members on documents, enhancing productivity and communication.",
      },
    ],
  },
  {
    icon: (
      <AnalyticsIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Reports",
    title: "Gain Insights and Analyze",
    overview:
      "Generate insightful reports to analyze project performance and make data-driven decisions.",
    breakdown: [
      {
        title: "Performance Metrics Analysis",
        overview:
          "Analyze project metrics to identify trends, measure performance, and pinpoint areas for improvement.",
      },
      {
        title: "Actionable Insights",
        overview:
          "Create actionable insights from project data to inform decision-making and strategic planning.",
      },
    ],
  },
  {
    icon: (
      <ChatIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Messaging",
    title: "Communicate and Collaborate",
    overview:
      "Engage in meaningful discussions through channels and direct messages, keeping everyone informed and aligned.",
    breakdown: [
      {
        title: "Team Communication",
        overview:
          "Engage in organized discussions with team members, sharing ideas, feedback, and updates.",
      },
      {
        title: "Alignment and Updates",
        overview:
          "Maintain effective communication to ensure all team members are informed and aligned with project goals and changes.",
      },
    ],
  },
  {
    icon: (
      <WhiteboardIcon
        className="relative -right-1 h-10 w-auto md:h-16"
        strokeWidth={1.3}
      />
    ),
    image: storyCard,
    name: "Whiteboards",
    title: "Brainstorm and Plan Visually",
    overview:
      "Brainstorm and plan with ease using interactive whiteboards, turning your ideas into actionable plans.",
    breakdown: [
      {
        title: "Collaborative Ideation",
        overview:
          "Brainstorm and visualize ideas with your team on digital whiteboards, enhancing creative thinking and planning.",
      },
      {
        title: "Interactive Planning",
        overview:
          "Use interactive whiteboard features to plan and organize project tasks and strategies effectively.",
      },
    ],
  },
];
