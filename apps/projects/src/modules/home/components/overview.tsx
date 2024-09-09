import { Box, Text, Flex, Wrapper } from "ui";

type CardProps = {
  title: string;
  count: number;
  description: string;
};
const Card = ({ title, count, description }: CardProps) => (
  <Wrapper>
    <Flex align="center" justify="between">
      <Text color="muted">{title}</Text>
      <svg
        className="h-5 w-auto text-primary"
        fill="none"
        height="24"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
        viewBox="0 0 24 24"
        width="24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path d="M7 7h10v10" />
        <path d="M7 17 17 7" />
      </svg>
    </Flex>
    <Text className="my-2" fontSize="2xl" fontWeight="semibold">
      {count}
    </Text>
    <Text color="muted">{description}</Text>
  </Wrapper>
);

export const Overview = () => {
  const overview = [
    {
      count: 27,
      description: "+5% from last month",
      title: "Stories closed",
    },
    {
      count: 8,
      description: "+2% from last month",
      title: "Stories overdue",
    },
    {
      count: 14,
      description: "+3% from last month",
      title: "Stories in progress",
    },
    {
      count: 5,
      description: "+1% from last month",
      title: "Created by you",
    },
    {
      count: 2,
      description: "+1% from last month",
      title: "Assigned to you",
    },
  ];
  return (
    <Box>
      <Text as="h2" fontSize="3xl" fontWeight="medium">
        Good afternoon, Joseph.
      </Text>
      <Text color="muted">
        Here&rsquo;s what&rsquo;s happening with your objectives today.
      </Text>

      <Box className="my-4 grid grid-cols-5 gap-4">
        {overview.map((item, i) => (
          <Card key={i} {...item} />
        ))}
      </Box>
    </Box>
  );
};
