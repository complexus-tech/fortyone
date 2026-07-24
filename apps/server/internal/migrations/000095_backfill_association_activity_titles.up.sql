UPDATE public.story_activities AS activity
SET current_value = related_story.title
FROM public.stories AS related_story
WHERE activity.reason IN (
        'association_added',
        'association_updated',
        'association_removed'
    )
    AND activity.field_changed IN (
        'blocked_by_id',
        'blocking_id',
        'related_id',
        'duplicate_id',
        'duplicated_by_id'
    )
    AND jsonb_typeof(activity.new_value) = 'string'
    AND activity.current_value = (activity.new_value #>> '{}')
    AND related_story.workspace_id = activity.workspace_id
    AND related_story.id::text = (activity.new_value #>> '{}');
