import { Box, Text, Button } from "ui";
import Link from "next/link";
import { Container } from "@/components/ui";

export const VoiceSection = () => {
  return (
    <Container className="py-20">
      <Box className="grid grid-cols-1 gap-16 md:grid-cols-2">
        <Box className="flex flex-col gap-6">
          <Text as="h2" className="text-5xl font-bold">
            We Value Your Voice
          </Text>
          <Text className="text-2xl">
            Giving should be a conversation, not a one-way transaction.
          </Text>
          <Box className="flex flex-col gap-4">
            <Text className="text-lg">
              That&apos;s why we invite donors to share feedback on:
            </Text>
            <ul className="list-disc pl-8 font-semibold">
              <li className="text-lg">Their experience using the platform</li>
              <li className="text-lg">
                The responsiveness of recipient organisations
              </li>
              <li className="text-lg">Suggestions for improvement</li>
            </ul>
          </Box>
          <Text className="text-lg">
            This feedback directly informs how we improve our systems and how we
            support the non-profits we host.
          </Text>
        </Box>

        <Box className="flex flex-col gap-6 rounded-lg p-8">
          <Box className="flex flex-col gap-4">
            <Button
              align="center"
              color="black"
              fullWidth
              href="/"
              variant="outline"
            >
              Linktree
            </Button>

            <Link href="/youtube">
              <Button className="w-full border border-gray-300 bg-transparent text-white hover:bg-white hover:text-secondary">
                YouTube
              </Button>
            </Link>
            <Link href="/substack">
              <Button className="w-full border border-gray-300 bg-transparent text-white hover:bg-white hover:text-secondary">
                Substack
              </Button>
            </Link>
            <Link href="/about-sivio">
              <Button className="w-full border border-gray-300 bg-transparent text-white hover:bg-white hover:text-secondary">
                Learn more about SIVIO
              </Button>
            </Link>
            <Link href="mailto:contact@africagiving.org">
              <Button className="w-full border border-gray-300 bg-transparent text-white hover:bg-white hover:text-secondary">
                Email Us
              </Button>
            </Link>
          </Box>

          <Box className="mt-auto flex justify-center gap-4">
            <Link href="/facebook">
              <Box className="hover:bg-gray-800 flex h-12 w-12 items-center justify-center rounded-full bg-black text-white">
                <Text className="text-lg font-bold">f</Text>
              </Box>
            </Link>
            <Link href="/instagram">
              <Box className="hover:bg-gray-800 flex h-12 w-12 items-center justify-center rounded-full bg-black text-white">
                <Text className="text-lg">📷</Text>
              </Box>
            </Link>
            <Link href="/twitter">
              <Box className="hover:bg-gray-800 flex h-12 w-12 items-center justify-center rounded-full bg-black text-white">
                <Text className="text-lg">🐦</Text>
              </Box>
            </Link>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
