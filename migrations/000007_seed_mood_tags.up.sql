INSERT INTO mood_tags (id, name, is_active)
VALUES
  ('34e969a4-a8e2-4935-b244-a8e78a3e70c6', 'calm', TRUE),
  ('bbf97c99-93c4-4bfc-95a6-4a34952ffde9', 'happy', TRUE),
  ('7d19ad4d-7218-4e8e-a192-b2e5258d381e', 'hopeful', TRUE),
  ('f9f8e53c-b337-4ea2-99f7-f07d0a6ec0ef', 'content', TRUE),
  ('bce3f089-3498-4695-bf4d-ac0c7b4b6baf', 'anxious', TRUE),
  ('5437c687-eb46-4e2f-b3ba-fb0373552274', 'stressed', TRUE),
  ('72d9983f-345c-4536-80fb-c6e5d25494cd', 'overwhelmed', TRUE),
  ('ea06b116-363a-41f6-b83a-a433ce9dfb9c', 'sad', TRUE),
  ('a3d2ad7a-1708-4eb5-ba6d-970f4f91af14', 'lonely', TRUE),
  ('96bbac8d-3c8f-453c-8276-55379454f723', 'tired', TRUE),
  ('a56486c9-6404-4d95-a2c2-1b8108fdf938', 'frustrated', TRUE),
  ('670174e0-6d52-457d-a109-18d4d3a0b32c', 'distracted', TRUE),
  ('df0162d1-f132-49b6-8601-02843eca465b', 'focused', TRUE),
  ('ec4d1d75-7ebb-4a6d-91be-b421fef09115', 'grateful', TRUE),
  ('6d1fe248-78be-43bc-bf2c-f66cd927fb80', 'numb', TRUE)
ON CONFLICT (name) DO NOTHING;
