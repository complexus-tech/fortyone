#!/usr/bin/env python3
"""Idempotent, loopback-only Chat demo data. Never sends notifications or runs AI."""
import argparse
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import subprocess
import uuid
from urllib.parse import urlparse, unquote

ROOT = Path(__file__).resolve().parents[1]
NAMESPACE = uuid.UUID("e7825289-5e29-437c-bf8c-3125b0d0cd27")


def sql(value):
    return "'" + str(value).replace("'", "''") + "'"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workspace", required=True, help="Existing local workspace slug")
    args = parser.parse_args()
    result = subprocess.run(["go", "run", "./internal/tools/databaseurl"], cwd=ROOT, text=True, capture_output=True)
    if result.returncode:
        raise SystemExit("Could not read the API database configuration")
    parsed = urlparse(result.stdout.strip())
    if parsed.hostname not in {"localhost", "127.0.0.1", "::1"}:
        raise SystemExit("Refusing to seed a non-loopback database")
    env = os.environ.copy()
    env.update(PGHOST=parsed.hostname, PGPORT=parsed.port and str(parsed.port) or "5432", PGUSER=unquote(parsed.username or ""), PGPASSWORD=unquote(parsed.password or ""), PGDATABASE=unquote(parsed.path[1:]), PGSSLMODE="disable")

    def query(statement):
        execution = subprocess.run(["psql", "-X", "-A", "-t", "-v", "ON_ERROR_STOP=1"], input=statement, text=True, env=env, capture_output=True)
        if execution.returncode:
            # SQL contains only fixtures and identifiers, never connection secrets.
            raise SystemExit(execution.stderr.strip())
        return execution.stdout.strip()

    row = query(f"""SELECT json_build_object('workspace', w.workspace_id, 'owner', w.created_by,
        'team', t.team_id, 'conversation', c.id)
        FROM workspaces w JOIN teams t ON t.workspace_id=w.workspace_id
        JOIN team_members tm ON tm.team_id=t.team_id AND tm.user_id=w.created_by
        JOIN chat_conversations c ON c.team_id=t.team_id AND c.workspace_id=w.workspace_id
        WHERE w.slug={sql(args.workspace)} AND w.deleted_at IS NULL
        ORDER BY t.created_at,t.team_id LIMIT 1;""")
    if not row:
        raise SystemExit("No existing workspace/team owned by an active member found; apply Chat migrations first")
    scope = json.loads(row)
    workspace, owner, team, conversation = (str(uuid.UUID(scope[key])) for key in ("workspace", "owner", "team", "conversation"))
    identity = lambda name: str(uuid.uuid5(NAMESPACE, workspace + ":" + name))
    people = [(identity("sarah"), "Sarah Chen"), (identity("daniel"), "Daniel Kim"), (identity("alex"), "Alex Morgan")]
    statements = ["BEGIN;", "SET LOCAL lock_timeout='5s';", "SET LOCAL statement_timeout='30s';"]
    for person, name in people:
        email = "chat-demo-" + person + "@example.invalid"
        statements.append(f"INSERT INTO users(user_id,username,email,full_name,is_active,is_system,timezone) VALUES ({sql(person)},{sql('chat-demo-'+person)},{sql(email)},{sql(name)},true,false,'Africa/Harare') ON CONFLICT(user_id) DO NOTHING;")
        statements.append(f"INSERT INTO workspace_members(workspace_id,user_id,role) VALUES ({sql(workspace)},{sql(person)},'member') ON CONFLICT DO NOTHING;")
        statements.append(f"INSERT INTO team_members(team_id,user_id) VALUES ({sql(team)},{sql(person)}) ON CONFLICT DO NOTHING;")
        statements.append(f"INSERT INTO notification_preferences(user_id,workspace_id,preferences) VALUES ({sql(person)},{sql(workspace)},'{json.dumps({'chat_message': {'email': False, 'in_app': True, 'push': False}})}') ON CONFLICT(user_id,workspace_id) DO NOTHING;")
    dm = identity("direct-sarah")
    low, high = sorted([owner, people[0][0]])
    # Canonical participant pair may already exist from manual testing.
    statements.append(f"INSERT INTO chat_conversations(id,workspace_id,kind,dm_user_low,dm_user_high) VALUES ({sql(dm)},{sql(workspace)},'dm',{sql(low)},{sql(high)}) ON CONFLICT(workspace_id,dm_user_low,dm_user_high) WHERE kind='dm' DO NOTHING;")
    statements.append(f"CREATE TEMP TABLE seed_chat_rooms ON COMMIT DROP AS SELECT id FROM chat_conversations WHERE workspace_id={sql(workspace)} AND (id={sql(conversation)} OR (kind='dm' AND dm_user_low={sql(low)} AND dm_user_high={sql(high)}));")
    # Serialize against live sends while assigning conversation-local sequences.
    statements.append("SELECT id FROM chat_conversations WHERE id IN (SELECT id FROM seed_chat_rooms) ORDER BY id FOR UPDATE;")
    messages = [
        "Morning team. Let’s review onboarding feedback before we pick this week’s priorities.",
        "I tested the latest flow. The team picker is much clearer, and the empty states explain the next step well.",
        "One detail to check: the invitation state should stay visible while the request is pending. We should also make the retry action easier to find.",
        "I’ve updated the acceptance criteria. Please keep keyboard navigation and smaller screens in the review.",
        "Release notes draft: https://example.com/release-notes — I’ve added the onboarding changes and the search improvements.",
        "The loading state feels better now. I’ll compare it on a slow connection this afternoon.",
        "Let’s keep the first release focused: clear conversations, reliable replies, and notifications that are useful.",
        "Design review notes: https://example.com/design-review — spacing and attachment previews are ready for feedback.",
        "Thanks for the review. I’ve captured the remaining edge cases and linked them to the release checklist.",
        "Ready for another pass. Try a long thread, then return to the main conversation and check your scroll position.",
        "The accessibility checks look good. We still need to verify focus after closing a thread and switching conversations.",
        "I’ll take the follow-up on search. The results should show enough context to identify the right message.",
    ]
    base = dt.datetime.now(dt.timezone.utc).replace(hour=8, minute=0, second=0, microsecond=0) - dt.timedelta(days=9)
    authors = [owner] + [person for person, _ in people]

    def message(key, room, author, body, timestamp, root=None):
        message_id, creation_id = identity(key), identity(key+":creation")
        digest = hashlib.sha256(("chat-demo-v1:"+key).encode()).hexdigest()
        content = json.dumps({"type": "doc", "content": [{"type": "paragraph", "content": [{"type": "text", "text": body}]}]})
        root_sql = sql(root) if root else "NULL"
        statements.append(f"""WITH next AS (UPDATE chat_conversations SET next_sequence=next_sequence+1
            WHERE id={room} AND NOT EXISTS(SELECT 1 FROM chat_conversation_messages WHERE id={sql(message_id)}) RETURNING id,next_sequence)
            INSERT INTO chat_conversation_messages(id,workspace_id,conversation_id,sequence,thread_root_id,author_id,creation_id,creation_hash,text,rich_content,created_at)
            SELECT {sql(message_id)},{sql(workspace)},next.id,next.next_sequence,{root_sql},{sql(author)},{sql(creation_id)},decode({sql(digest)},'hex'),{sql(body)},CAST({sql(content)} AS jsonb),{sql(timestamp.isoformat())} FROM next;""")
        if "https://" in body:
            link = next(part for part in body.split() if part.startswith("https://"))
            statements.append(f"INSERT INTO chat_message_links(message_id,url,title) VALUES ({sql(message_id)},{sql(link)},'Team reference') ON CONFLICT(message_id,url) DO NOTHING;")
        return message_id

    for index in range(180):
        stamp = base + dt.timedelta(days=index//18, minutes=(index%18)*21)
        root = message(f"team:{index}", sql(conversation), authors[index%4], messages[index%len(messages)], stamp)
        if index%9 == 0:
            for reply in range(3):
                message(f"team:{index}:reply:{reply}", sql(conversation), authors[(index+reply+1)%4], ["Agreed. I’ll check this in the next pass.", "I’ve added a note so we can track that decision.", "Thanks, this is ready for review."][reply], stamp+dt.timedelta(minutes=reply+1),root)
        if index%7 == 0:
            statements.append(f"INSERT INTO chat_message_reactions(message_id,user_id,emoji) VALUES ({sql(root)},{sql(people[0][0])},'👍') ON CONFLICT DO NOTHING;")
    room = f"(SELECT id FROM chat_conversations WHERE workspace_id={sql(workspace)} AND kind='dm' AND dm_user_low={sql(low)} AND dm_user_high={sql(high)})"
    for index in range(12):
        message(f"dm:{index}",room,authors[index%2],["Are you available to review the onboarding flow?", "Yes, I’ll leave feedback in the team conversation.", "The latest checklist is ready. Let’s compare the empty states."][index%3],base+dt.timedelta(days=8,minutes=index*15))
    statements.append("UPDATE chat_conversations c SET last_message_at=(SELECT max(created_at) FROM chat_conversation_messages m WHERE m.conversation_id=c.id) WHERE c.id IN (SELECT id FROM seed_chat_rooms);")
    statements.append("COMMIT;")
    query("\n".join(statements))
    print(f"Seeded {args.workspace}: 180 team messages, 60 replies, 12 direct messages, 3 demo members, links and reactions. Re-running is safe. No notification events or Maya requests were created.")


if __name__ == "__main__":
    main()
