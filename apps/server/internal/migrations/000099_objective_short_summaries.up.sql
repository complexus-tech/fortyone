ALTER TABLE public.objectives
ADD COLUMN short_summary varchar(500);

UPDATE public.objectives AS objective
SET short_summary = summaries.short_summary
FROM (
    VALUES
        ('6fc4fccc-0bb9-4299-baa6-afadc47de6b2'::uuid, 'Complete and validate the work defined for this objective.'),
        ('d787b45a-5f03-4e93-b1da-413ba52e7802'::uuid, 'Build a super admin dashboard for platform-wide oversight and management.'),
        ('6ce000a8-16ac-44f7-bcaf-fe2e49fecf76'::uuid, 'Build the APIs trade authorities need to onboard suppliers and oversee national commerce activity.'),
        ('1cc4fb12-a10c-4095-8fe3-207cb98195a2'::uuid, 'Establish the backend, database, authentication, access control, and shared services that all other work depends on.'),
        ('1c694ffa-cab1-442c-956b-9a0b0c1da34d'::uuid, 'Launch an MVP that helps art fairs manage ticketing, invitations, event discovery, payments, and analytics.'),
        ('13f1a435-b0e7-4ff2-83b3-cffebc3b76f3'::uuid, 'Increase overall output by improving delivery focus and execution.'),
        ('be1e128a-cbf0-48c7-a023-b5eaf2290e72'::uuid, 'Advance development of the core platform and its foundational capabilities.'),
        ('2660113f-89c4-42b2-afc3-60ca07ab7b33'::uuid, 'Implement a clearer team structure that improves execution efficiency.'),
        ('d104b1a2-73f3-41a9-aad0-4d29b4271ee2'::uuid, 'Restructure the team to improve ownership, collaboration, and delivery efficiency.'),
        ('5217db47-fd4d-4a8f-8684-54ec913ae0a8'::uuid, 'Deploy the current product changes to a stable production environment.'),
        ('9579245e-9779-4896-a064-3b27d1756c14'::uuid, 'Complete the Q3 2028 team restructure to improve execution efficiency.'),
        ('a958e217-9eac-4876-aa69-e8589b0266f5'::uuid, 'Validate the objective workflow through a focused test initiative.'),
        ('1bd0fec2-9c5f-4d41-9d6b-27596c868039'::uuid, 'Deploy to production, onboard beta testers, and grow the user base from 60,000 to 90,000.'),
        ('225b9302-9bb0-47ce-a8b6-6e8f422adbc8'::uuid, 'Complete the Q3 2028 team restructure quickly to improve execution efficiency.'),
        ('909df22a-9832-4bb5-8bc2-f78245166661'::uuid, 'Deploy the current product changes to a stable production environment.'),
        ('39ba7eec-1d3d-4126-9c35-a06101403624'::uuid, 'Make customer onboarding faster, clearer, and more seamless.'),
        ('bc39985d-17e8-43ec-9779-df36a72d6ea6'::uuid, 'Deploy AI capabilities into the product and make them available to users.'),
        ('02d7e478-a100-4ec3-a955-dd6b21653c73'::uuid, 'Complete the Q3 2029 team restructure to improve execution efficiency.'),
        ('d794a662-c3cc-421e-8021-f9b707d2c093'::uuid, 'Launch One Fund and complete the work required for a successful release.'),
        ('b610aac5-95d6-4232-a8b9-d4baefbc7991'::uuid, 'Design a polished FortyOne demo for a prospective client.'),
        ('a2d81ac4-cc38-434f-9046-656c4f56dc09'::uuid, 'Increase revenue by 35% through focused, measurable growth initiatives.'),
        ('b7cf1ebd-c873-4e58-91d4-2bd6c58370d2'::uuid, 'Generate and qualify two new sales leads.'),
        ('e996e436-4883-4f07-92ba-2ecc07aab657'::uuid, 'Publish the mobile app and make it available to users.'),
        ('08471a9b-c737-4a9b-80fc-c9cc82e90278'::uuid, 'Create and execute a social media strategy that grows awareness and engagement.'),
        ('207c7814-dd03-4574-802b-d7628f8563c9'::uuid, 'Launch the Games On The Go website for customers and partners.'),
        ('b292013b-b02f-45e2-a9c3-e79c2daac0a3'::uuid, 'Complete the application design within two weeks.'),
        ('8f4ee563-6f2b-49db-be14-83b8aa7855c7'::uuid, 'Launch Go Ticket and make the ticketing experience available to customers.'),
        ('8656bc35-2ae0-46b3-83af-fc916d50025f'::uuid, 'Build an ongoing second-brain system for learning, knowledge, and personal development.'),
        ('a7ff4aa8-fe23-418c-8016-e91524b90c25'::uuid, 'Maintain and improve health-related skills, routines, and long-term wellbeing.'),
        ('2025a9c1-4cb5-4ebd-8ebf-96aa783164ab'::uuid, 'Build and maintain a clear public presence across professional, creative, and social channels.'),
        ('fc1c383c-a934-42f9-aef0-185fddf8de45'::uuid, 'Tăng doanh thu lên 20% thông qua các hoạt động tăng trưởng có thể đo lường.'),
        ('371819a4-1b8c-417c-b9e0-6c8f83a156dc'::uuid, 'Ra mắt sản phẩm mới và đưa sản phẩm đến khách hàng mục tiêu.'),
        ('6f5ecbc4-ec87-471e-befb-4b4eb167ac20'::uuid, 'Organize and present the skills needed to secure a mid-level software role earning more than US$78,000 annually.'),
        ('2f4907a6-9631-44f9-b1bb-b56a18b8d48e'::uuid, 'Increase sales by 30% through focused acquisition and conversion improvements.'),
        ('26b271cd-f3ea-4628-9e66-c66d008b20eb'::uuid, 'Increase completed objectives by 22% through stronger planning and execution.'),
        ('2fa8be13-499a-4bf5-8066-f3051311ff6b'::uuid, 'Complete and deliver the work defined for Yosub.'),
        ('cab539e5-9f31-4c9f-9438-ccc8918bcd5b'::uuid, 'Reduce cloud infrastructure costs by 50% without compromising reliability.'),
        ('1f1e7daf-c4bd-47a4-b8c9-3780f8e11ea3'::uuid, 'Expand and refine the product interface to improve usability and visual quality.'),
        ('522a25ae-07f9-49d8-8ccf-cd2becdc4789'::uuid, 'Validate the objective workflow with a focused test objective.'),
        ('c68c5ee6-ab58-4dab-a166-865f9c0cfa9a'::uuid, 'Complete WePOS compliance work in collaboration with the accountant.')
) AS summaries(objective_id, short_summary)
WHERE objective.objective_id = summaries.objective_id;

UPDATE public.objectives
SET short_summary = LEFT(TRIM(name), 500)
WHERE short_summary IS NULL;
