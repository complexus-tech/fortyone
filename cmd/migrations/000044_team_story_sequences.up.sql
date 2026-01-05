-- 000044_team_story_sequences.up.sql
CREATE TABLE public.team_story_sequences (
    team_id uuid NOT NULL,
    last_sequence_id integer DEFAULT 0 NOT NULL
);

ALTER TABLE ONLY public.team_story_sequences ADD CONSTRAINT team_story_sequences_pkey PRIMARY KEY (team_id);

ALTER TABLE ONLY public.team_story_sequences 
    ADD CONSTRAINT team_story_sequences_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;
