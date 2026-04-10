INSERT INTO users (id, name, email, password)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'Test User',
  'test@example.com',
  '$2a$12$Ay5WiIR4EYiNESRaoIPcTuhm/nUwnGphYOGiEpOO7.KvxRT25tkR2'
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id)
VALUES (
  '22222222-2222-2222-2222-222222222222',
  'Demo Project',
  'Seeded project for local testing',
  '11111111-1111-1111-1111-111111111111'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tasks (
  id, title, description, status, priority,
  project_id, assignee_id, created_by, due_date
)
VALUES
(
  '33333333-3333-3333-3333-333333333333',
  'First task',
  'Seeded task in todo',
  'todo',
  'low',
  '22222222-2222-2222-2222-222222222222',
  NULL,
  '11111111-1111-1111-1111-111111111111',
  now() + interval '7 days'
),
(
  '44444444-4444-4444-4444-444444444444',
  'Second task',
  'Seeded task in progress',
  'in_progress',
  'medium',
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  '11111111-1111-1111-1111-111111111111',
  now() + interval '14 days'
),
(
  '55555555-5555-5555-5555-555555555555',
  'Third task',
  'Seeded task done',
  'done',
  'high',
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  '11111111-1111-1111-1111-111111111111',
  NULL
)
ON CONFLICT (id) DO NOTHING;

