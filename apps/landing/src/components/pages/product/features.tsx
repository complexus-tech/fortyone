import { Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";
import { Blur } from "@/components/ui";
import {
  AnalyticsIcon,
  ChatIcon,
  DocsIcon,
  EpicsIcon,
  MilestonesIcon,
  ObjectiveIcon,
  RoadmapIcon,
  StoryIcon,
  WhiteboardIcon,
} from "icons";
import storyCard from "../../../../public/kanban.png";
import Image from "next/image";

export const Features = () => {
  const features = [
    {
      icon: (
        <StoryIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Stories",
      title: "Track your work",
      overview: "Track and manage tasks, issues, and bugs with user stories.",
      breakdown: [
        {
          title: "Issue Properties",
          overview:
            "Add priority, labels, estimates, issue relations, references, and due dates to stories.",
        },
        {
          title: "Optimized for Speed",
          overview:
            "Quickly create, edit, and manage stories without relying on mouse interactions.",
        },
        {
          title: "Parent and Sub-stories",
          overview:
            "Break down larger tasks into smaller pieces or group related stories together.",
        },
        {
          title: "Story Templates",
          overview:
            "Utilize predefined templates for faster story creation and consistent information sharing.",
        },
        {
          title: "Detect Similar Stories",
          overview:
            "Use AI to identify possible duplicate stories and maintain workspace cleanliness.",
        },
        {
          title: "Story Importer",
          overview:
            "Effortlessly import stories from external issue trackers for seamless migration.",
        },
      ],
    },
    {
      icon: (
        <ObjectiveIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Objectives",
      title: "Set goals and track progress",
      overview: "Define clear objectives to guide your project's direction.",
      breakdown: [
        {
          title: "Clear Objectives",
          overview:
            "Define specific and achievable goals to guide the project's direction.",
        },
        {
          title: "Progress Tracking",
          overview:
            "Monitor and track the progress of objectives to ensure alignment with project goals.",
        },
      ],
    },
    {
      icon: (
        <RoadmapIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Roadmaps",
      title: "Plan and visualize project roadmaps",
      overview:
        "Create and visualize project roadmaps to plan and track progress.",
      breakdown: [
        {
          title: "Visual Planning",
          overview:
            "Visualize project milestones and timelines for effective planning.",
        },
        {
          title: "Progress Tracking",
          overview:
            "Track project progress and milestones to ensure timely delivery.",
        },
      ],
    },
    {
      icon: (
        <MilestonesIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Sprints",
      title: "Iterative development and delivery",
      overview:
        "Organize work into sprints for iterative development and delivery.",
      breakdown: [
        {
          title: "Iterative Development",
          overview:
            "Break down project tasks into manageable sprints for continuous improvement.",
        },
        {
          title: "Delivery Management",
          overview:
            "Manage and track the delivery of features and functionalities within sprints.",
        },
      ],
    },
    {
      icon: (
        <EpicsIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Epics",
      title: "Manage large-scale features or initiatives",
      overview: "Manage and track large-scale features or initiatives.",
      breakdown: [
        {
          title: "Feature Management",
          overview:
            "Manage and track the progress of large-scale features or initiatives.",
        },
        {
          title: "Initiative Tracking",
          overview: "Track the progress and alignment of project initiatives.",
        },
      ],
    },
    {
      icon: (
        <DocsIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Documents",
      title: "Document management",
      overview: "Store and collaborate on project-related documents.",
      breakdown: [
        {
          title: "Document Storage",
          overview:
            "Store and organize project-related documents for easy access and collaboration.",
        },
        {
          title: "Collaboration Tools",
          overview:
            "Collaborate with team members in real-time on project documentation.",
        },
      ],
    },
    {
      icon: (
        <AnalyticsIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Reports",
      title: "Analytics and insights",
      overview: "Generate reports to analyze project performance and metrics.",
      breakdown: [
        {
          title: "Performance Analysis",
          overview:
            "Analyze project performance metrics to identify trends and areas for improvement.",
        },
        {
          title: "Insight Generation",
          overview:
            "Generate actionable insights from project data to inform decision-making.",
        },
      ],
    },
    {
      icon: (
        <ChatIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Discussions",
      title: "Collaborate and communicate",
      overview:
        "Engage in meaningful discussions and keep everyone informed and aligned.",
      breakdown: [
        {
          title: "Collaborative Discussions",
          overview:
            "Engage in discussions with team members to share ideas, feedback, and updates.",
        },
        {
          title: "Alignment Communication",
          overview:
            "Keep team members informed and aligned through effective communication.",
        },
      ],
    },
    {
      icon: (
        <WhiteboardIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      image: storyCard,
      name: "Whiteboards",
      title: "Visualize ideas and plans",
      overview:
        "Visualize ideas and plans effortlessly with advanced whiteboard functionality.",
      breakdown: [
        {
          title: "Visual Ideation",
          overview:
            "Brainstorm and visualize ideas collaboratively on digital whiteboards.",
        },
        {
          title: "Planning Tools",
          overview:
            "Plan and organize project tasks and strategies using interactive whiteboard features.",
        },
      ],
    },
  ];

  return (
    <Container className="max-w-4xl">
      {features.map(
        ({ name, title, overview, icon, breakdown, image }, idx) => (
          <Box
            key={name}
            id={name.toLowerCase()}
            className="relative scroll-mt-12 border-t border-gray-200/5 pt-8 md:py-12"
          >
            <Flex
              gap={6}
              justify="between"
              align="center"
              className="relative -left-1 mb-2 opacity-15 md:mb-6"
            >
              <Text
                fontWeight="bold"
                className="text-stroke-white text-[2.6rem] md:text-6xl"
              >
                <span className="hidden md:inline">0{idx + 1}. </span>
                {name}
              </Text>
              {icon}
            </Flex>
            <Box className="col-span-3">
              <Text
                fontWeight="normal"
                className="text-xl opacity-80 md:text-3xl"
              >
                {title}
              </Text>
              <Text
                fontWeight="normal"
                fontSize="lg"
                color="muted"
                className="my-2 md:my-4"
              >
                {overview}
              </Text>
              <Box className="rounded-xl bg-dark-100 p-1 md:p-1.5">
                <Image
                  placeholder="blur"
                  className="pointer-events-none rounded-lg"
                  src={image}
                  alt={name}
                />
              </Box>
              <Box className="mt-6 px-0.5">
                {breakdown.map(({ title, overview }) => (
                  <Box key={title} className="">
                    <Text fontSize="lg" className="mb-2">
                      {title}
                    </Text>
                    <Text color="muted" fontWeight="normal" className="mb-5">
                      {overview}
                    </Text>
                  </Box>
                ))}
              </Box>
            </Box>
            <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[650px] w-[650px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
          </Box>
        ),
      )}
      <Text
        fontWeight="normal"
        className="mb-16 mt-12 text-xl leading-snug opacity-80 md:mb-28 md:mt-0 md:text-2xl"
      >
        Complexus is tailored to equip you with comprehensive tools for
        effective project management. Whether you're tracking tasks, visualizing
        project roadmaps, or generating reports, Complexus has everything you
        need to succeed.
      </Text>
    </Container>
  );
};
