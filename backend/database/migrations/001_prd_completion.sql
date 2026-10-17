-- Additive migration for existing SIKAP-J installations.
ALTER TABLE sikapj.users
  ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE sikapj.e_berkas
SET status_verifikasi = 'Menunggu Verifikasi',
    catatan_admin = NULL,
    verified_at = NULL,
    verified_by = NULL
WHERE status_verifikasi <> 'Menunggu Verifikasi'
  AND (verified_at IS NULL OR verified_by IS NULL
       OR (status_verifikasi = 'Ditolak' AND NULLIF(BTRIM(catatan_admin), '') IS NULL));

CREATE TABLE IF NOT EXISTS sikapj.alert_log (
  id BIGSERIAL PRIMARY KEY,
  nip VARCHAR(18) NOT NULL REFERENCES sikapj.pegawai(nip) ON DELETE CASCADE,
  tipe_alert VARCHAR(3) NOT NULL CHECK (tipe_alert IN ('KGB','KP')),
  tanggal_jatuh_tempo DATE NOT NULL,
  status VARCHAR(20) NOT NULL CHECK (status IN ('MENDEKATI','SEGERA','TERTUTUP')),
  notified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (nip, tipe_alert, tanggal_jatuh_tempo)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pegawai_user_unique
  ON sikapj.pegawai(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_pegawai_active_name
  ON sikapj.pegawai(is_active, nama_lengkap);
CREATE INDEX IF NOT EXISTS idx_berkas_status_uploaded
  ON sikapj.e_berkas(status_verifikasi, uploaded_at);
CREATE INDEX IF NOT EXISTS idx_kinerja_nip_period
  ON sikapj.arsip_kinerja(nip, tahun_periode DESC);
CREATE INDEX IF NOT EXISTS idx_diklat_nip_date
  ON sikapj.riwayat_diklat(nip, tanggal_mulai DESC);
CREATE INDEX IF NOT EXISTS idx_audit_created
  ON sikapj.audit_log(created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action_created
  ON sikapj.audit_log(action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alert_status_due
  ON sikapj.alert_log(status, tanggal_jatuh_tempo);

DO $$ BEGIN
  ALTER TABLE sikapj.pegawai ADD CONSTRAINT pegawai_nip_format_ck CHECK (nip ~ '^[0-9]{18}$');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
  ALTER TABLE sikapj.e_berkas ADD CONSTRAINT e_berkas_file_size_ck CHECK (file_size_bytes > 0);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
  ALTER TABLE sikapj.e_berkas ADD CONSTRAINT e_berkas_rejection_note_ck CHECK (status_verifikasi <> 'Ditolak' OR NULLIF(BTRIM(catatan_admin), '') IS NOT NULL);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
  ALTER TABLE sikapj.riwayat_diklat ADD CONSTRAINT riwayat_diklat_jp_ck CHECK (jumlah_jp > 0);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
  ALTER TABLE sikapj.riwayat_diklat ADD CONSTRAINT riwayat_diklat_dates_ck CHECK (tanggal_selesai >= tanggal_mulai);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
