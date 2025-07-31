import { Box, Text, Button } from "ui";
import Link from "next/link";
import { FacebookIcon, InstagramIcon, TwitterIcon } from "icons";

export const VoiceSection = () => {
  return (
    <Box className="py-20">
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
              rounded="none"
              variant="outline"
            >
              Linktree
            </Button>

            <Button
              align="center"
              color="black"
              fullWidth
              href="/"
              rounded="none"
              variant="outline"
            >
              YouTube
            </Button>
            <Link href="/substack">
              <Button
                align="center"
                color="black"
                fullWidth
                href="/"
                rounded="none"
                variant="outline"
              >
                Substack
              </Button>
            </Link>
            <Link href="/about-sivio">
              <Button
                align="center"
                color="black"
                fullWidth
                href="/"
                rounded="none"
                variant="outline"
              >
                Learn more about SIVIO
              </Button>
            </Link>
            <Link href="mailto:contact@africagiving.org">
              <Button
                align="center"
                color="black"
                fullWidth
                href="/"
                rounded="none"
                variant="outline"
              >
                Email Us
              </Button>
            </Link>
          </Box>

          <Box className="mt-8 flex justify-center gap-4">
            <Link href="/facebook">
              <FacebookIcon className="text-black" />
            </Link>
            <Link href="/instagram">
              <InstagramIcon className="text-black" />
            </Link>
            <Link href="/twitter">
              <TwitterIcon className="text-black" />
            </Link>
          </Box>
        </Box>
      </Box>
    </Box>
  );
};
