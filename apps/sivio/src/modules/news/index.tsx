import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { NewsCard } from "./components/news-card";

const news = [
  {
    id: 1,
    title: "AfricaGiving Launches New Platform",
    description: "Text",
    image: "/images/news/launch.jpg",
    link: "/news/launch",
  },
  {
    id: 2,
    title: "Community Impact Report 2024",
    description: "Text",
    image: "/images/news/impact-2024.jpg",
    link: "/news/impact-2024",
  },
  {
    id: 3,
    title: "Partnership Announcement with SIVIO",
    description: "Text",
    image: "/images/news/partnership.jpg",
    link: "/news/partnership",
  },
  {
    id: 4,
    title: "Transparency Initiative Launch",
    description: "Text",
    image: "/images/news/transparency.jpg",
    link: "/news/transparency",
  },
  {
    id: 5,
    title: "Annual Donor Recognition Event",
    description: "Text",
    image: "/images/news/donor-event.jpg",
    link: "/news/donor-event",
  },
  {
    id: 6,
    title: "Technology Platform Updates",
    description: "Text",
    image: "/images/news/tech-updates.jpg",
    link: "/news/tech-updates",
  },
];

export const NewsPage = () => {
  return (
    <Container className="pb-28 pt-12">
      <Text as="h1" className="mb-12 text-5xl font-semibold text-black">
        Events
      </Text>
      <Box className="grid grid-cols-1 gap-12 md:grid-cols-3">
        {news.map((item) => (
          <NewsCard key={item.id} news={item} />
        ))}
      </Box>
    </Container>
  );
};
