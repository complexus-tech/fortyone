ALTER TABLE public.story_comment_reactions
    DROP CONSTRAINT story_comment_reactions_emoji_check,
    ADD CONSTRAINT story_comment_reactions_emoji_check
        CHECK (emoji IN ('👍', '✅', '👀', '❤️', '🎉', '😂', '🙏', '🚀', '💡', '🔥', '👏', '🙌', '🤔', '😊', '😢', '👎', '💯'));
