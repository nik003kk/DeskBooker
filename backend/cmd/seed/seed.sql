-- Repeatable development fixtures; deterministic IDs and upsert-on-conflict.
-- Demo identities are fictional, and no credentials/passwords are stored.
INSERT INTO demo_seed.users (id, full_name, email, role) VALUES
  (1, 'Ana Novak', 'ana.novak@example.test', 'admin'),
  (2, 'Marko Kranjc', 'marko.kranjc@example.test', 'manager'),
  (3, 'Nina Zajc', 'nina.zajc@example.test', 'employee'),
  (4, 'Luka Horvat', 'luka.horvat@example.test', 'employee')
ON CONFLICT (id) DO UPDATE SET full_name = EXCLUDED.full_name, email = EXCLUDED.email, role = EXCLUDED.role;

INSERT INTO demo_seed.desks (id, code, floor, zone, has_monitor) VALUES
  (1, 'D-101', 1, 'A', TRUE),
  (2, 'D-102', 1, 'A', FALSE),
  (3, 'D-201', 2, 'B', TRUE),
  (4, 'D-202', 2, 'B', TRUE),
  (5, 'D-203', 2, 'B', FALSE)
ON CONFLICT (id) DO UPDATE SET code = EXCLUDED.code, floor = EXCLUDED.floor, zone = EXCLUDED.zone, has_monitor = EXCLUDED.has_monitor;

INSERT INTO demo_seed.meeting_rooms (id, name, floor, capacity) VALUES
  (1, 'Triglav', 1, 8),
  (2, 'Pohorje', 2, 4),
  (3, 'Bled', 2, 12)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, floor = EXCLUDED.floor, capacity = EXCLUDED.capacity;

INSERT INTO demo_seed.parking_spaces (id, code, covered, electric_charging) VALUES
  (1, 'P-01', TRUE, FALSE),
  (2, 'P-02', TRUE, TRUE),
  (3, 'P-03', FALSE, FALSE),
  (4, 'P-04', FALSE, TRUE)
ON CONFLICT (id) DO UPDATE SET code = EXCLUDED.code, covered = EXCLUDED.covered, electric_charging = EXCLUDED.electric_charging;

INSERT INTO demo_seed.equipment (id, name, category, quantity) VALUES
  (1, 'Dell 27-inch monitor', 'monitor', 10),
  (2, 'USB-C docking station', 'dock', 6),
  (3, 'Portable projector', 'projector', 2),
  (4, 'Video conference camera', 'camera', 3)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, category = EXCLUDED.category, quantity = EXCLUDED.quantity;
