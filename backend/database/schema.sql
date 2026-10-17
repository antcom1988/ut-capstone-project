BEGIN;
CREATE SCHEMA IF NOT EXISTS sikapj;

CREATE TABLE IF NOT EXISTS sikapj.users (
 id BIGSERIAL PRIMARY KEY, username VARCHAR(50) NOT NULL UNIQUE, password_hash VARCHAR(255) NOT NULL,
 role VARCHAR(20) NOT NULL CHECK (role IN ('ADMIN','PEGAWAI')), is_active BOOLEAN NOT NULL DEFAULT TRUE,
 must_change_password BOOLEAN NOT NULL DEFAULT TRUE, last_login TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE sikapj.users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE sikapj.users ALTER COLUMN must_change_password SET DEFAULT TRUE;

CREATE TABLE IF NOT EXISTS sikapj.pegawai (
 nip VARCHAR(18) PRIMARY KEY CHECK (nip ~ '^[0-9]{18}$'), user_id BIGINT UNIQUE REFERENCES sikapj.users(id) ON DELETE SET NULL,
 nama_lengkap VARCHAR(150) NOT NULL, tempat_lahir VARCHAR(100), tanggal_lahir DATE, jenis_kelamin VARCHAR(1) CHECK(jenis_kelamin IN('L','P')),
 email VARCHAR(254), no_hp VARCHAR(20), status_pegawai VARCHAR(10) NOT NULL CHECK(status_pegawai IN('PNS','PPPK')),
 bidang VARCHAR(100) NOT NULL, jenis_jabatan VARCHAR(30) NOT NULL CHECK(jenis_jabatan IN('Struktural','Fungsional','Pelaksana')),
 nama_jabatan VARCHAR(150) NOT NULL, pangkat VARCHAR(50) NOT NULL, golongan VARCHAR(10) NOT NULL,
 tmt_pengangkatan DATE NOT NULL, tmt_pangkat_terakhir DATE NOT NULL, tmt_kgb_terakhir DATE NOT NULL,
 jenjang_pendidikan VARCHAR(20), jurusan VARCHAR(100), institusi_pendidikan VARCHAR(150), is_active BOOLEAN NOT NULL DEFAULT TRUE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sikapj.master_dokumen (
 id INTEGER PRIMARY KEY, nama_dokumen VARCHAR(100) NOT NULL, max_size_mb INTEGER NOT NULL DEFAULT 1 CHECK(max_size_mb BETWEEN 1 AND 20),
 is_required BOOLEAN NOT NULL DEFAULT TRUE, deskripsi VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS sikapj.e_berkas (
 id BIGSERIAL PRIMARY KEY, nip VARCHAR(18) NOT NULL REFERENCES sikapj.pegawai(nip) ON DELETE CASCADE,
 dokumen_id INTEGER NOT NULL REFERENCES sikapj.master_dokumen(id), file_path VARCHAR(500) NOT NULL, file_name VARCHAR(255) NOT NULL,
 file_size_bytes BIGINT NOT NULL CHECK(file_size_bytes>0), mime_type VARCHAR(50) NOT NULL,
 status_verifikasi VARCHAR(30) NOT NULL DEFAULT 'Menunggu Verifikasi' CHECK(status_verifikasi IN('Menunggu Verifikasi','Disetujui','Ditolak')),
 catatan_admin TEXT, uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), verified_at TIMESTAMPTZ, verified_by BIGINT REFERENCES sikapj.users(id),
 UNIQUE(nip,dokumen_id), CHECK(status_verifikasi<>'Ditolak' OR NULLIF(BTRIM(catatan_admin),'') IS NOT NULL)
);
CREATE TABLE IF NOT EXISTS sikapj.arsip_kinerja (
 id BIGSERIAL PRIMARY KEY,nip VARCHAR(18) NOT NULL REFERENCES sikapj.pegawai(nip) ON DELETE CASCADE,tahun_periode INTEGER NOT NULL CHECK(tahun_periode BETWEEN 2000 AND 2200),
 jenis_dokumen VARCHAR(10) NOT NULL CHECK(jenis_dokumen IN('PK','SKP')),file_path VARCHAR(500) NOT NULL,file_name VARCHAR(255) NOT NULL,
 file_size_bytes BIGINT NOT NULL CHECK(file_size_bytes>0),uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),UNIQUE(nip,tahun_periode,jenis_dokumen)
);
CREATE TABLE IF NOT EXISTS sikapj.riwayat_diklat (
 id BIGSERIAL PRIMARY KEY,nip VARCHAR(18) NOT NULL REFERENCES sikapj.pegawai(nip) ON DELETE CASCADE,nama_diklat VARCHAR(200) NOT NULL,
 penyelenggara VARCHAR(150) NOT NULL,jumlah_jp INTEGER NOT NULL CHECK(jumlah_jp>0),tanggal_mulai DATE NOT NULL,tanggal_selesai DATE NOT NULL,
 file_sertifikat_path VARCHAR(500),file_name VARCHAR(255),created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),CHECK(tanggal_selesai>=tanggal_mulai)
);
CREATE TABLE IF NOT EXISTS sikapj.settings(key VARCHAR(50) PRIMARY KEY,value VARCHAR(255) NOT NULL,deskripsi VARCHAR(255),updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE IF NOT EXISTS sikapj.audit_log(id BIGSERIAL PRIMARY KEY,user_id BIGINT REFERENCES sikapj.users(id) ON DELETE SET NULL,action VARCHAR(50) NOT NULL,details JSONB,ip_address INET,created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE IF NOT EXISTS sikapj.alert_log(
 id BIGSERIAL PRIMARY KEY,nip VARCHAR(18) NOT NULL REFERENCES sikapj.pegawai(nip) ON DELETE CASCADE,tipe_alert VARCHAR(3) NOT NULL CHECK(tipe_alert IN('KGB','KP')),
 tanggal_jatuh_tempo DATE NOT NULL,status VARCHAR(20) NOT NULL CHECK(status IN('MENDEKATI','SEGERA','TERTUTUP')),notified_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),UNIQUE(nip,tipe_alert,tanggal_jatuh_tempo)
);

CREATE INDEX IF NOT EXISTS idx_pegawai_active_name ON sikapj.pegawai(is_active,nama_lengkap);
CREATE INDEX IF NOT EXISTS idx_berkas_status_uploaded ON sikapj.e_berkas(status_verifikasi,uploaded_at);
CREATE INDEX IF NOT EXISTS idx_berkas_nip ON sikapj.e_berkas(nip);
CREATE INDEX IF NOT EXISTS idx_kinerja_nip_period ON sikapj.arsip_kinerja(nip,tahun_periode DESC);
CREATE INDEX IF NOT EXISTS idx_diklat_nip_date ON sikapj.riwayat_diklat(nip,tanggal_mulai DESC);
CREATE INDEX IF NOT EXISTS idx_audit_created ON sikapj.audit_log(created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action_created ON sikapj.audit_log(action,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alert_status_due ON sikapj.alert_log(status,tanggal_jatuh_tempo);

INSERT INTO sikapj.settings(key,value,deskripsi) VALUES('ALERT_N_MONTHS','3','Ambang pengingat KGB/KP dalam bulan') ON CONFLICT(key) DO NOTHING;
INSERT INTO sikapj.master_dokumen(id,nama_dokumen,max_size_mb,is_required,deskripsi) VALUES
(1,'KTP',1,TRUE,'Kartu Tanda Penduduk'),(2,'KK',1,TRUE,'Kartu Keluarga'),(3,'SK CPNS',1,TRUE,'SK pengangkatan CPNS'),
(4,'SK PNS/PPPK',1,TRUE,'SK pengangkatan PNS/PPPK'),(5,'SK Kenaikan Pangkat Terakhir',1,TRUE,'SK pangkat terakhir'),
(6,'SK Jabatan',1,TRUE,'SK jabatan'),(7,'SKP 2 Tahun Sebelumnya',2,TRUE,'SKP dua tahun sebelumnya'),
(8,'SKP 1 Tahun Terakhir',1,TRUE,'SKP tahun terakhir'),(9,'SK Jabatan Fungsional',1,FALSE,'Wajib untuk jabatan fungsional'),
(10,'Ijazah Pendidikan Terakhir',1,TRUE,'Ijazah terakhir'),(11,'Transkrip Nilai',1,TRUE,'Transkrip terakhir'),
(12,'Akreditasi Universitas/PT',1,TRUE,'Bukti akreditasi')
ON CONFLICT(id) DO UPDATE SET nama_dokumen=EXCLUDED.nama_dokumen,max_size_mb=EXCLUDED.max_size_mb,is_required=EXCLUDED.is_required,deskripsi=EXCLUDED.deskripsi;
COMMIT;
