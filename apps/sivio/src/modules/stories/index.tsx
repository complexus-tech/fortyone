import { Box } from "ui";
import { Container } from "@/components/ui";
import { StoryCard } from "./components/story-card";
import { Hero } from "./components/hero";

const stories = [
  {
    id: 1,
    title: "Impact Story: Education in Rural Kenya",
    description: "Text",
    image: "/images/stories/kenya-education.jpg",
    link: "/stories/kenya-education",
  },
  {
    id: 2,
    title: "Community Transformation in Ghana",
    description: "Text",
    image: "/images/stories/ghana-community.jpg",
    link: "/stories/ghana-community",
  },
  {
    id: 3,
    title: "Women Empowerment in Nigeria",
    description: "Text",
    image: "/images/stories/nigeria-women.jpg",
    link: "/stories/nigeria-women",
  },
  {
    id: 4,
    title: "Youth Development in South Africa",
    description: "Text",
    image: "/images/stories/south-africa-youth.jpg",
    link: "/stories/south-africa-youth",
  },
  {
    id: 5,
    title: "Healthcare Access in Tanzania",
    description: "Text",
    image: "/images/stories/tanzania-healthcare.jpg",
    link: "/stories/tanzania-healthcare",
  },
  {
    id: 6,
    title: "Agricultural Innovation in Uganda",
    description: "Text",
    image: "/images/stories/uganda-agriculture.jpg",
    link: "/stories/uganda-agriculture",
  },
];

export const StoriesPage = () => {
  return (
    <Box>
      <Hero />
      <Container className="pb-28 pt-16">
        <Box className="grid grid-cols-1 gap-12 md:grid-cols-3">
          {stories.map((story) => (
            <StoryCard key={story.id} story={story} />
          ))}
        </Box>
      </Container>
    </Box>
  );
};
