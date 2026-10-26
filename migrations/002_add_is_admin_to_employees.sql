-- 002: penanda super admin per karyawan (Opsi B, 26 Okt 2026).
-- Kolom ini juga dibuat otomatis oleh GORM AutoMigrate saat aplikasi start
-- (migrateAndSeed di internal/config/config.go); file ini dokumentasi +
-- jalur penerapan manual bila diperlukan.
ALTER TABLE employees ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- Pengangkatan bendahara (diterapkan terpisah via psql SESUDAH deploy,
-- karena kolom baru ada setelah AutoMigrate berjalan):
-- UPDATE employees SET siakad_id = 'cmqerhl790001x75oylsmz5gl', is_admin = TRUE
-- WHERE id = '220df786-ce22-4a57-92eb-2952fcacd82d'; -- Karismaning Ulfa. A, S.Pd (= Ustadzah Karisma, Siakad username 'karisma')
