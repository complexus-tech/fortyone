import { BlurImage, Box, Flex, Text } from "ui";
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

export const Features = () => {
  const features = [
    {
      icon: (
        <StoryIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Stories",
      title: "Track your work",
      overview: "Track and manage tasks, issues, and bugs with user stories.",
    },
    {
      icon: (
        <ObjectiveIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Objectives",
      title: "Set goals and track progress",
      overview: "Define clear objectives to guide your project's direction.",
    },
    {
      icon: (
        <RoadmapIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Roadmaps",
      title: "Plan and visualize project roadmaps",
      overview:
        "Create and visualize project roadmaps to plan and track progress.",
    },
    {
      icon: (
        <MilestonesIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Sprints",
      title: "Iterative development and delivery",
      overview:
        "Organize work into sprints for iterative development and delivery.",
    },
    {
      icon: (
        <EpicsIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Epics",
      title: "Manage large-scale features or initiatives",
      overview: "Manage and track large-scale features or initiatives.",
    },
    {
      icon: (
        <DocsIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Documents",
      title: "Document management",
      overview: "Store and collaborate on project-related documents.",
    },
    {
      icon: (
        <AnalyticsIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Reports",
      title: "Analytics and insights",
      overview: "Generate reports to analyze project performance and metrics.",
    },
    {
      icon: (
        <ChatIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Discussions",
      title: "Collaborate and communicate",
      overview:
        "Engage in meaningful discussions and keep everyone informed and aligned.",
    },
    {
      icon: (
        <WhiteboardIcon
          strokeWidth={1.3}
          className="relative -right-1 h-10 w-auto md:h-16"
        />
      ),
      name: "Whiteboards",
      title: "Visualize ideas and plans",
      overview:
        "Visualize ideas and plans effortlessly with advanced whiteboard functionality.",
    },
  ];

  return (
    <Container className="max-w-4xl">
      {features.map(({ name, title, overview, icon }, idx) => (
        <Box
          key={name}
          id={name.toLowerCase()}
          className="relative scroll-mt-12 border-t border-gray-200/5 pt-8 md:py-12"
        >
          <Flex
            gap={6}
            justify="between"
            align="center"
            className="relative -left-1 mb-6 opacity-15"
          >
            <Text
              fontWeight="bold"
              className="text-stroke-white text-5xl md:text-6xl"
            >
              <span className="hidden md:inline">0{idx + 1}. </span>
              {name}
            </Text>
            {icon}
          </Flex>
          <Box className="col-span-3">
            <Text fontSize="3xl" fontWeight="normal" className="opacity-80">
              {title}
            </Text>
            <Text
              fontWeight="normal"
              fontSize="lg"
              color="muted"
              className="my-4"
            >
              {overview}
            </Text>
            <BlurImage
              src="/stories.png"
              theme="dark"
              objectPosition="bottom"
              className="aspect-[5/3]"
            />
          </Box>
          <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[650px] w-[650px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
        </Box>
      ))}
      <Text
        fontSize="2xl"
        fontWeight="normal"
        className="mb-28 mt-10 leading-snug opacity-80"
      >
        Complexus is tailored to equip you with comprehensive tools for
        effective project management. Whether you're tracking tasks, visualizing
        project roadmaps, or generating reports, Complexus has everything you
        need to succeed.
      </Text>
    </Container>
  );
};
