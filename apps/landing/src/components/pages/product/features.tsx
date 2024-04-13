import { BlurImage, Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";
import { Blur } from "@/components/ui";
import { ObjectiveIcon } from "icons";

export const Features = () => {
  const features = [
    {
      name: "Stories",
      title: "Track your work",
      description:
        "Track and manage tasks, issues, and bugs with user stories.",
    },
    {
      name: "Objectives",
      title: "Set goals and track progress",
      description: "Define clear objectives to guide your project's direction.",
    },
    {
      name: "Roadmaps",
      title: "Plan and visualize project roadmaps",
      description:
        "Create and visualize project roadmaps to plan and track progress.",
    },
    {
      name: "Sprints",
      title: "Iterative development and delivery",
      description:
        "Organize work into sprints for iterative development and delivery.",
    },
    {
      name: "Epics",
      title: "Manage large-scale features or initiatives",
      description: "Manage and track large-scale features or initiatives.",
    },
    {
      name: "Documents",
      title: "Document management",
      description: "Store and collaborate on project-related documents.",
    },
    {
      name: "Reports",
      title: "Analytics and insights",
      description:
        "Generate reports to analyze project performance and metrics.",
    },
    {
      name: "Discussions",
      title: "Collaborate and communicate",
      description:
        "Engage in meaningful discussions and keep everyone informed and aligned.",
    },
    {
      name: "Whiteboards",
      title: "Visualize ideas and plans",
      description:
        "Visualize ideas and plans effortlessly with advanced whiteboard functionality.",
    },
  ];

  return (
    <Container className="max-w-4xl">
      {features.map(({ name, title, description }, idx) => (
        <Box
          key={name}
          id={name.toLowerCase()}
          className="relative border-t border-gray-200/5 pt-8 md:py-12"
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
            <ObjectiveIcon
              strokeWidth={1.2}
              className="relative -right-1 h-10 w-auto md:h-16"
            />
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
              {description}
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
