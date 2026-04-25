-- ============================================
-- REMOVE DEFAULT PROCESS SEEDS
-- ============================================

DELETE FROM transitions WHERE process_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM statuses WHERE process_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM processes WHERE id = '00000000-0000-0000-0000-000000000001';