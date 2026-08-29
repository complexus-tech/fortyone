UPDATE story_activities AS activity
SET current_value = key_result.name
FROM key_results AS key_result
WHERE activity.field_changed = 'key_result_id'
  AND activity.current_value = key_result.id::text;
