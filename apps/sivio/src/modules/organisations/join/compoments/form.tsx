import { Box, Button, Text, Input, TextArea, Checkbox } from "ui";
import { Container } from "@/components/ui";

export const Form = () => {
  return (
    <Container className="max-w-4xl py-10">
      <Text className="mb-10 text-4xl font-semibold">
        Application to join the Africa Giving platform
      </Text>

      <form className="space-y-8">
        {/* Organisation Information */}
        <Box className="space-y-4">
          <Text className="text-xl font-semibold">
            Organisation Information
          </Text>

          <Box>
            <Text className="mb-2 font-medium">1. Organisation name*</Text>
            <Input
              className="w-full border-dark"
              placeholder="Enter your organisation name"
              required
              type="text"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">2. Organisation type*</Text>
            <select className="w-full rounded-lg border border-dark px-3 py-2 text-sm focus:border-primary focus:outline-none">
              <option value="">Select organisation type</option>
              <option value="ngo">NGO</option>
              <option value="cbo">Community-based organisation</option>
              <option value="fbo">Faith-based organisation</option>
              <option value="other">Other</option>
            </select>
          </Box>

          <Box>
            <Text className="mb-2 font-medium">3. Country of operation*</Text>
            <select className="w-full rounded-lg border border-dark px-3 py-2 text-sm focus:border-primary focus:outline-none">
              <option value="">Select country</option>
              <option value="kenya">Kenya</option>
              <option value="ghana">Ghana</option>
              <option value="nigeria">Nigeria</option>
              <option value="south-africa">South Africa</option>
              <option value="tanzania">Tanzania</option>
              <option value="uganda">Uganda</option>
              <option value="other">Other</option>
            </select>
          </Box>

          <Box>
            <Text className="mb-2 font-medium">4. Year established</Text>
            <Input
              className="w-full border-dark"
              max="2024"
              min="1900"
              placeholder="e.g., 2010"
              type="number"
            />
          </Box>
        </Box>

        {/* Contact Information */}
        <Box className="space-y-4">
          <Text className="text-xl font-semibold">Contact Information</Text>

          <Box>
            <Text className="mb-2 font-medium">5. Primary contact name*</Text>
            <Input
              className="w-full border-dark"
              placeholder="Enter contact name"
              required
              type="text"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">6. Primary contact email*</Text>
            <Input
              className="w-full border-dark"
              placeholder="Enter contact email"
              required
              type="email"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">7. Primary contact phone*</Text>
            <Input
              className="w-full border-dark"
              placeholder="Enter contact phone"
              required
              type="tel"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">8. Website</Text>
            <Input
              className="w-full border-dark"
              placeholder="https://your-website.com"
              type="url"
            />
          </Box>
        </Box>

        {/* Mission & Focus */}
        <Box className="space-y-4">
          <Text className="text-xl font-semibold">Mission & Focus</Text>

          <Box>
            <Text className="mb-2 font-medium">9. Mission statement*</Text>
            <TextArea
              className="w-full border-dark"
              placeholder="Describe your organisation's mission and goals"
              required
              rows={4}
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">10. Primary focus areas*</Text>
            <Box className="space-y-2">
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Education</Text>
              </Box>
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Healthcare</Text>
              </Box>
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Youth empowerment</Text>
              </Box>
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Women empowerment</Text>
              </Box>
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Community development</Text>
              </Box>
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Environmental conservation</Text>
              </Box>
              <Box className="flex items-center gap-2">
                <Checkbox />
                <Text>Other</Text>
              </Box>
            </Box>
          </Box>
        </Box>

        {/* Financial & Operational */}
        <Box className="space-y-4">
          <Text className="text-xl font-semibold">Financial & Operational</Text>

          <Box>
            <Text className="mb-2 font-medium">11. Annual budget range*</Text>
            <select
              className="w-full rounded-lg border border-dark px-3 py-2 text-sm focus:border-primary focus:outline-none"
              required
            >
              <option value="">Select budget range</option>
              <option value="under-10k">Under $10,000</option>
              <option value="10k-50k">$10,000 - $50,000</option>
              <option value="50k-100k">$50,000 - $100,000</option>
              <option value="100k-500k">$100,000 - $500,000</option>
              <option value="over-500k">Over $500,000</option>
            </select>
          </Box>

          <Box>
            <Text className="mb-2 font-medium">
              12. Number of beneficiaries served
            </Text>
            <Input
              className="w-full border-dark"
              min="0"
              placeholder="Enter number of beneficiaries"
              type="number"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">13. Geographic scope*</Text>
            <select
              className="w-full rounded-lg border border-dark px-3 py-2 text-sm focus:border-primary focus:outline-none"
              required
            >
              <option value="">Select geographic scope</option>
              <option value="local">Local community</option>
              <option value="regional">Regional</option>
              <option value="national">National</option>
              <option value="international">International</option>
            </select>
          </Box>
        </Box>

        {/* Documentation */}
        <Box className="space-y-4">
          <Text className="text-xl font-semibold">Documentation</Text>

          <Box>
            <Text className="mb-2 font-medium">
              14. Registration certificate*
            </Text>
            <Input
              className="w-full border-dark"
              placeholder="Upload registration certificate"
              required
              type="text"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">15. Financial statements</Text>
            <Input
              className="w-full border-dark"
              placeholder="Upload financial statements"
              type="text"
            />
          </Box>

          <Box>
            <Text className="mb-2 font-medium">16. Annual report</Text>
            <Input
              className="w-full border-dark"
              placeholder="Upload annual report"
              type="text"
            />
          </Box>
        </Box>

        {/* Additional Information */}
        <Box className="space-y-4">
          <Text className="text-xl font-semibold">Additional Information</Text>

          <Box>
            <Text className="mb-2 font-medium">
              17. How did you hear about AfricaGiving?
            </Text>
            <select className="w-full rounded-lg border border-dark px-3 py-2 text-sm focus:border-primary focus:outline-none">
              <option value="">Select option</option>
              <option value="social-media">Social media</option>
              <option value="referral">Referral</option>
              <option value="search">Search engine</option>
              <option value="conference">Conference/event</option>
              <option value="other">Other</option>
            </select>
          </Box>
        </Box>

        <Button className="mx-auto" size="lg" type="submit">
          Submit Application
        </Button>
      </form>
    </Container>
  );
};
