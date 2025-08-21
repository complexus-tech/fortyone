import { Flex, Text, Box, Badge } from "ui";
import { Container } from "@/components/ui";

// Custom icons for the capabilities
const ReasoningIcon = () => (
  <svg
    className="h-8 w-8 text-dark dark:text-white"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    viewBox="0 0 24 24"
  >
    <path
      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

const VisionIcon = () => (
  <svg
    className="h-8 w-8 text-dark dark:text-white"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    viewBox="0 0 24 24"
  >
    <path
      d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

const ToolCallingIcon = () => (
  <svg
    className="h-8 w-8 text-dark dark:text-white"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    viewBox="0 0 24 24"
  >
    <rect height="18" rx="2" ry="2" width="18" x="3" y="3" />
    <path d="M9 12l3 3 3-3" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

const StructuredOutputsIcon = () => (
  <svg
    className="h-8 w-8 text-dark dark:text-white"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    viewBox="0 0 24 24"
  >
    <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
    <path d="M14 2v6h6" />
    <path d="M16 13H8" />
    <path d="M16 17H8" />
    <path d="M10 9H8" />
  </svg>
);

const ImageGenerationIcon = () => (
  <svg
    className="h-8 w-8 text-dark dark:text-white"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    viewBox="0 0 24 24"
  >
    <rect height="18" rx="2" ry="2" width="18" x="3" y="3" />
    <circle cx="8.5" cy="8.5" r="1.5" />
    <path d="M21 15l-5-5L5 21" />
  </svg>
);

const SearchIcon = () => (
  <svg
    className="h-8 w-8 text-dark dark:text-white"
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    viewBox="0 0 24 24"
  >
    <path
      d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

type Capability = {
  icon: React.ComponentType;
  title: string;
  description: string;
  isNew?: boolean;
};

const capabilities: Capability[] = [
  {
    icon: ReasoningIcon,
    title: "Reasoning",
    description:
      "Understand the universe: solve humanity's most difficult scientific problems with deep thought.",
  },
  {
    icon: VisionIcon,
    title: "Vision",
    description:
      "See the world through vision, interpreting images and visuals with sharp, insightful understanding.",
  },
  {
    icon: ToolCallingIcon,
    title: "Tool calling",
    description:
      "Harness external power with tool calling, seamlessly integrating third-party functions.",
  },
  {
    icon: StructuredOutputsIcon,
    title: "Structured outputs",
    description:
      "Organize chaos with structured outputs, delivering clean, predictable responses.",
  },
  {
    icon: ImageGenerationIcon,
    title: "Image generation",
    description:
      "Bring your ideas to life with image generation, creating visuals that are as unique as you are.",
  },
  {
    icon: SearchIcon,
    title: "Search",
    description:
      "Tap into the now with real-time search, pulling fresh, relevant data from the web and X instantly.",
    isNew: true,
  },
];

const CapabilityCard = ({ capability }: { capability: Capability }) => {
  const IconComponent = capability.icon;

  return (
    <Box className="group border border-transparent border-l-gray-100 bg-gradient-to-b px-7 py-9 hover:border-gray-100 hover:from-gray-50 dark:border-l-dark-100 dark:hover:border-dark-100 dark:hover:from-dark-300">
      <Flex align="center" className="mb-5" justify="between">
        <IconComponent />
        {capability.isNew ? (
          <Badge className="bg-blue-600/20 text-blue-400 border-blue-600/30">
            New
          </Badge>
        ) : null}
      </Flex>
      <Text as="h3" className="mb-4 text-lg dark:text-white">
        {capability.title}
      </Text>
      <Text className="text-[0.95rem] leading-relaxed opacity-60">
        {capability.description}
      </Text>
    </Box>
  );
};

export const CoreValues = () => {
  return (
    <Container className="pb-28 pt-12">
      <Text className="mb-8 text-sm uppercase tracking-wider opacity-70">
        [ CAPABILITIES ]
      </Text>
      <Text as="h2" className="text-4xl font-semibold md:text-5xl">
        Models that fit your needs
      </Text>

      <Box className="mt-24 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {capabilities.map((capability, index) => (
          <CapabilityCard capability={capability} key={index} />
        ))}
      </Box>
    </Container>
  );
};
