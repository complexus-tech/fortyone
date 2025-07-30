import { BlurImage, Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";

export const Content = () => {
  return (
    <Container className="py-10">
      <Box className="grid grid-cols-3 gap-16">
        <BlurImage
          alt="Transparency"
          className="aspect-square"
          src="/images/transparency/1.png"
        />
        <Box className="col-span-2">
          <Flex className="gap-6" direction="column">
            <Text className="text-lg">
              AfricaGiving is committed to ensuring that all organisations
              featured on our platform meet essential standards of legitimacy,
              transparency, and accountability. To be listed, organisations must
              be legally registered in their country of operation and must
              operate as non-profit organisations. Our verification process is
              designed to build trust with donors, partners, and the communities
              these organisations serve.
            </Text>

            <Text className="text-lg">
              Organisations undergo a structured evaluation based on five key
              areas. Each criterion contributes to an overall compliance score,
              which reflects the organisation&apos;s level of transparency and
              reliability.
            </Text>

            <Flex className="gap-6" direction="column">
              <Box>
                <Text
                  as="h3"
                  className="mb-3 text-3xl font-semibold text-black"
                >
                  Legal Registration
                </Text>
                <ul className="list-disc pl-8">
                  <li>
                    <Text className="mb-2 text-lg">
                      Organisations must provide proof of official registration
                      with the relevant authorities in their country.
                    </Text>
                  </li>
                  <li>
                    <Text className="text-lg">
                      Verification documents may include certificates of
                      incorporation, NGO registration papers, or other
                      government-issued credentials.
                    </Text>
                  </li>
                </ul>
              </Box>

              <Box>
                <Text
                  as="h3"
                  className="mb-3 text-3xl font-semibold text-black"
                >
                  Financial Transparency
                </Text>
                <Text className="mb-2 text-lg">
                  Financial accountability is crucial for donor confidence.
                  Organisations are assessed on their ability to share:
                </Text>
                <Flex className="gap-1" direction="column">
                  <ul className="list-disc pl-8">
                    <li>
                      <Text className="text-lg">
                        Audited financial statements
                      </Text>
                      <ul className="list-disc pl-8">
                        <li>
                          <Text className="text-lg">Annual reports</Text>
                        </li>
                        <li>
                          <Text className="text-lg">
                            Basic financial disclosures detailing annual budget
                            goals
                          </Text>
                        </li>
                      </ul>
                    </li>
                  </ul>
                </Flex>
              </Box>

              <Box>
                <Text
                  as="h3"
                  className="mb-3 text-3xl font-semibold text-black"
                >
                  Governance Structures
                </Text>
                <Text className="mb-2 text-lg">
                  Strong governance ensures responsible oversight and ethical
                  decision-making. Organisations must demonstrate:
                </Text>
                <Flex className="gap-1" direction="column">
                  <Text className="text-lg">
                    • A functioning board of directors or advisory group
                    overseeing operations.
                  </Text>
                  <Text className="text-lg">
                    • Established leadership structures that support
                    accountability and strategic direction.
                  </Text>
                </Flex>
              </Box>

              <Box>
                <Text
                  as="h3"
                  className="mb-3 text-3xl font-semibold text-black"
                >
                  Reporting Practices
                </Text>
                <Text className="mb-2 text-lg">
                  Organisations must consistently communicate their activities,
                  impact, and use of funds. This includes:
                </Text>
                <Flex className="gap-1" direction="column">
                  <Text className="text-lg">• Project reports</Text>
                  <Text className="text-lg">• Impact statements</Text>
                  <Text className="text-lg">
                    • Regular updates shared with donors and stakeholders.
                  </Text>
                </Flex>
              </Box>

              <Box>
                <Text
                  as="h3"
                  className="mb-3 text-3xl font-semibold text-black"
                >
                  Accessibility & Public Presence
                </Text>
                <Text className="mb-2 text-lg">
                  To be visible and accessible to donors, organisations must
                  have:
                </Text>
                <Flex className="gap-1" direction="column">
                  <Text className="text-lg">
                    • A functional website, email, or social media presence to
                    engage with supporters.
                  </Text>
                  <Text className="text-lg">
                    • Clear channels for communication and transparency.
                  </Text>
                </Flex>
              </Box>
            </Flex>

            <Text className="text-lg">
              Once assessed, organisations receive a compliance score based on
              their level of adherence to these criteria. High-scoring
              organisations signal strong transparency and accountability, while
              those with lower scores may be asked to enhance specific aspects
              before approval.
            </Text>

            <Text className="text-lg">
              AfricaGiving ensures that every organisation meets our standards,
              empowering donors to give confidently and supporting non-profits
              in fostering meaningful change.
            </Text>

            <Box className="mt-6">
              <Text className="text-lg">
                Before being listed on AfricaGiving, organisations must review
                and sign a Memorandum of Understanding (MOU) with SIVIO
                Institute, the parent organisation overseeing the platform. This
                MOU serves as a formal agreement outlining the terms and
                conditions of use, ensuring clarity on roles, responsibilities,
                and expectations for all listed entities.
              </Text>
            </Box>

            <Text className="text-lg">
              To streamline the onboarding process, organisations must sign the
              MOU digitally before their profiles become active. This ensures
              efficiency, security, and ease of access while reinforcing a
              commitment to AfricaGiving&apos;s principles.
            </Text>

            <Box className="mt-6">
              <Text className="text-lg">
                Should you have any further questions, you can reach out to the
                AfricaGiving team:{" "}
                <span className="font-bold underline">
                  africagiving@sivioinstitute.org
                </span>
              </Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Container>
  );
};
