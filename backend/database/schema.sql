create table if not exists users
(
    id            serial
        primary key,
    username      varchar(50)  not null
        unique,
    password_hash varchar(255) not null,
    role          varchar(20)  not null
        constraint users_role_check
            check ((role)::text = ANY ((ARRAY ['ADMIN'::character varying, 'PEGAWAI'::character varying])::text[])),
    is_active     boolean                  default true,
    last_login    timestamp with time zone,
    created_at    timestamp with time zone default CURRENT_TIMESTAMP,
    updated_at    timestamp with time zone default CURRENT_TIMESTAMP
);

create table if not exists pegawai
(
    nip                  varchar(18)  not null
        primary key,
    user_id              integer
                                      references users
                                          on delete set null,
    nama_lengkap         varchar(150) not null,
    tempat_lahir         varchar(100),
    tanggal_lahir        date,
    jenis_kelamin        varchar(1)
        constraint pegawai_jenis_kelamin_check
            check ((jenis_kelamin)::text = ANY ((ARRAY ['L'::character varying, 'P'::character varying])::text[])),
    email                varchar(100),
    no_hp                varchar(20),
    status_pegawai       varchar(10)  not null
        constraint pegawai_status_pegawai_check
            check ((status_pegawai)::text = ANY
                   ((ARRAY ['PNS'::character varying, 'PPPK'::character varying])::text[])),
    bidang               varchar(100) not null,
    jenis_jabatan        varchar(30)  not null
        constraint pegawai_jenis_jabatan_check
            check ((jenis_jabatan)::text = ANY
                   ((ARRAY ['Struktural'::character varying, 'Fungsional'::character varying, 'Pelaksana'::character varying])::text[])),
    nama_jabatan         varchar(150) not null,
    pangkat              varchar(50)  not null,
    golongan             varchar(10)  not null,
    tmt_pengangkatan     date         not null,
    tmt_pangkat_terakhir date         not null,
    tmt_kgb_terakhir     date         not null,
    jenjang_pendidikan   varchar(20),
    jurusan              varchar(100),
    institusi_pendidikan varchar(150),
    is_active            boolean                  default true,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP
);

create table if not exists master_dokumen
(
    id           serial
        primary key,
    nama_dokumen varchar(100)      not null,
    max_size_mb  integer default 1 not null,
    is_required  boolean default true,
    deskripsi    varchar(255)
);

create table if not exists e_berkas
(
    id                serial
        primary key,
    nip               varchar(18)  not null
        references pegawai
            on delete cascade,
    dokumen_id        integer      not null
        references master_dokumen,
    file_path         varchar(255) not null,
    file_name         varchar(255) not null,
    file_size_bytes   bigint       not null,
    mime_type         varchar(50)  not null,
    status_verifikasi varchar(30)              default 'Menunggu Verifikasi'::character varying
        constraint e_berkas_status_verifikasi_check
            check ((status_verifikasi)::text = ANY
                   ((ARRAY ['Menunggu Verifikasi'::character varying, 'Disetujui'::character varying, 'Ditolak'::character varying])::text[])),
    catatan_admin     text,
    uploaded_at       timestamp with time zone default CURRENT_TIMESTAMP,
    verified_at       timestamp with time zone,
    verified_by       integer
        references users,
    constraint unique_pegawai_dokumen
        unique (nip, dokumen_id)
);

create table if not exists arsip_kinerja
(
    id              serial
        primary key,
    nip             varchar(18)  not null
        references pegawai
            on delete cascade,
    tahun_periode   integer      not null,
    jenis_dokumen   varchar(10)  not null
        constraint arsip_kinerja_jenis_dokumen_check
            check ((jenis_dokumen)::text = ANY ((ARRAY ['PK'::character varying, 'SKP'::character varying])::text[])),
    file_path       varchar(255) not null,
    file_name       varchar(255) not null,
    file_size_bytes bigint       not null,
    uploaded_at     timestamp with time zone default CURRENT_TIMESTAMP,
    constraint unique_pegawai_periode_jenis
        unique (nip, tahun_periode, jenis_dokumen)
);

create table if not exists riwayat_diklat
(
    id                   serial
        primary key,
    nip                  varchar(18)                        not null
        references pegawai
            on delete cascade,
    nama_diklat          varchar(200)                       not null,
    penyelenggara        varchar(150)                       not null,
    jumlah_jp            integer                  default 0 not null,
    tanggal_mulai        date                               not null,
    tanggal_selesai      date                               not null,
    file_sertifikat_path varchar(255),
    file_name            varchar(255),
    created_at           timestamp with time zone default CURRENT_TIMESTAMP
);

create table if not exists settings
(
    key        varchar(50)  not null
        primary key,
    value      varchar(255) not null,
    deskripsi  varchar(255),
    updated_at timestamp with time zone default CURRENT_TIMESTAMP
);

create table if not exists audit_log
(
    id         serial
        primary key,
    user_id    integer
                           references users
                               on delete set null,
    action     varchar(50) not null,
    details    text,
    ip_address varchar(50),
    created_at timestamp with time zone default CURRENT_TIMESTAMP
);

INSERT INTO sikapj.pegawai (nip, user_id, nama_lengkap, tempat_lahir, tanggal_lahir, jenis_kelamin, email, no_hp, status_pegawai, bidang, jenis_jabatan, nama_jabatan, pangkat, golongan, tmt_pengangkatan, tmt_pangkat_terakhir, tmt_kgb_terakhir, jenjang_pendidikan, jurusan, institusi_pendidikan, is_active, created_at, updated_at) VALUES ('199003202014022005', 3, 'Siti Aminah, S.A.P.', 'Surabaya', '1990-03-20', 'P', 'siti.aminah@jemberkab.go.id', '081298765432', 'PNS', 'Bidang Anggaran', 'Pelaksana', 'Pengelola Sistem Informasi Keuangan', 'Penata Muda', 'III/a', '2014-02-01', '2022-04-01', '2024-04-01', 'S1', 'Administrasi Publik', 'Universitas Brawijaya', true, '2026-10-09 13:53:01.218132 +00:00', '2026-10-09 13:53:01.218132 +00:00');
INSERT INTO sikapj.pegawai (nip, user_id, nama_lengkap, tempat_lahir, tanggal_lahir, jenis_kelamin, email, no_hp, status_pegawai, bidang, jenis_jabatan, nama_jabatan, pangkat, golongan, tmt_pengangkatan, tmt_pangkat_terakhir, tmt_kgb_terakhir, jenjang_pendidikan, jurusan, institusi_pendidikan, is_active, created_at, updated_at) VALUES ('198811102012011003', 4, 'Rudi Hartono, S.Kom.', 'Banyuwangi', '1988-11-10', 'L', 'rudi.hartono@jemberkab.go.id', '081345678901', 'PNS', 'Sekretariat', 'Fungsional', 'Pranata Komputer Ahli Pertama', 'Penata Muda Tingkat I', 'III/b', '2012-01-01', '2023-04-01', '2025-03-01', 'S1', 'Sistem Informasi', 'Universitas Terbuka', true, '2026-10-09 13:53:01.218132 +00:00', '2026-10-09 13:53:01.218132 +00:00');
INSERT INTO sikapj.pegawai (nip, user_id, nama_lengkap, tempat_lahir, tanggal_lahir, jenis_kelamin, email, no_hp, status_pegawai, bidang, jenis_jabatan, nama_jabatan, pangkat, golongan, tmt_pengangkatan, tmt_pangkat_terakhir, tmt_kgb_terakhir, jenjang_pendidikan, jurusan, institusi_pendidikan, is_active, created_at, updated_at) VALUES ('198507152010011012', 2, 'Budi Santoso, S.E., M.Ak.', 'Jakarta', '1985-07-15', 'L', 'budi.santoso@jemberkab.go.id', '081234567891', 'PNS', 'Bidang Akuntansi', 'Fungsional', 'Analis Keuangan Pusat dan Daerah Ahli Muda', 'Penata Tingkat I', 'III/d', '2010-01-01', '2022-10-01', '2024-12-01', 'S2', 'Magister Akuntansi', 'Universitas Jember', true, '2026-10-09 13:53:01.218132 +00:00', '2026-10-09 14:38:16.697513 +00:00');


INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (1, 'KTP', 1, true, 'Kartu Tanda Penduduk');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (2, 'KK', 1, true, 'Kartu Keluarga');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (3, 'SK CPNS', 1, true, 'Surat Keputusan Pengangkatan CPNS');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (4, 'SK PNS/PPPK', 1, true, 'Surat Keputusan Pengangkatan PNS/PPPK');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (5, 'SK Kenaikan Pangkat Terakhir', 1, true, 'Surat Keputusan Pangkat Terakhir');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (6, 'SK Jabatan', 1, true, 'Surat Keputusan Pengangkatan dalam Jabatan');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (7, 'SKP 2 Tahun Sebelumnya', 2, true, 'Sasaran Kinerja Pegawai 2 tahun sebelumnya (Maks 2 MB)');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (8, 'SKP 1 Tahun Terakhir', 1, true, 'Sasaran Kinerja Pegawai tahun terakhir');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (9, 'SK Jabatan Fungsional', 1, false, 'SK Jabatan Fungsional (bila menduduki jabatan fungsional)');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (10, 'Ijazah Pendidikan Terakhir', 1, true, 'Ijazah pendidikan formal terakhir');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (11, 'Transkrip Nilai', 1, true, 'Transkrip nilai pendidikan formal terakhir');
INSERT INTO sikapj.master_dokumen (id, nama_dokumen, max_size_mb, is_required, deskripsi) VALUES (12, 'Akreditasi Universitas/PT', 1, true, 'Bukti akreditasi perguruan tinggi/program studi');


INSERT INTO sikapj.arsip_kinerja (id, nip, tahun_periode, jenis_dokumen, file_path, file_name, file_size_bytes, uploaded_at) VALUES (1, '198507152010011012', 2025, 'SKP', 'uploads/kinerja/skp_2025_budi.pdf', 'skp_2025_budi.pdf', 850230, '2026-10-09 13:53:01.359502 +00:00');
INSERT INTO sikapj.arsip_kinerja (id, nip, tahun_periode, jenis_dokumen, file_path, file_name, file_size_bytes, uploaded_at) VALUES (2, '198507152010011012', 2026, 'PK', 'uploads/kinerja/pk_2026_budi.pdf', 'pk_2026_budi.pdf', 345000, '2026-10-09 13:53:01.359502 +00:00');

INSERT INTO sikapj.e_berkas (id, nip, dokumen_id, file_path, file_name, file_size_bytes, mime_type, status_verifikasi, catatan_admin, uploaded_at, verified_at, verified_by) VALUES (2, '198507152010011012', 2, 'uploads/berkas/kk_budi.pdf', 'kk_budi.pdf', 450000, 'application/pdf', 'Disetujui', null, '2026-10-09 13:53:01.449266 +00:00', null, null);
INSERT INTO sikapj.e_berkas (id, nip, dokumen_id, file_path, file_name, file_size_bytes, mime_type, status_verifikasi, catatan_admin, uploaded_at, verified_at, verified_by) VALUES (1, '198507152010011012', 1, 'uploads/berkas/ktp_budi.pdf', 'ktp_budi.pdf', 312000, 'application/pdf', 'Disetujui', '', '2026-10-09 13:53:01.449266 +00:00', '2026-10-09 14:26:23.692138 +00:00', 1);
INSERT INTO sikapj.e_berkas (id, nip, dokumen_id, file_path, file_name, file_size_bytes, mime_type, status_verifikasi, catatan_admin, uploaded_at, verified_at, verified_by) VALUES (3, '198507152010011012', 5, 'uploads/berkas/sk_pangkat_budi.pdf', 'sk_pangkat_budi.pdf', 890000, 'application/pdf', 'Ditolak', 'salah', '2026-10-09 13:53:01.449266 +00:00', '2026-10-09 14:27:44.281268 +00:00', 1);
INSERT INTO sikapj.e_berkas (id, nip, dokumen_id, file_path, file_name, file_size_bytes, mime_type, status_verifikasi, catatan_admin, uploaded_at, verified_at, verified_by) VALUES (5, '199003202014022005', 2, 'storage\uploads\berkas\199003202014022005\2_2568f20f.pdf', 'Billing Registrasi Matakuliah -.pdf', 138716, 'application/pdf', 'Ditolak', 'dokumen salah', '2026-10-09 14:30:01.963883 +00:00', '2026-10-09 14:34:22.299011 +00:00', 1);
INSERT INTO sikapj.e_berkas (id, nip, dokumen_id, file_path, file_name, file_size_bytes, mime_type, status_verifikasi, catatan_admin, uploaded_at, verified_at, verified_by) VALUES (4, '199003202014022005', 1, 'storage\uploads\berkas\199003202014022005\1_c145259f.png', 'ektm (1).png', 163244, 'image/png', 'Disetujui', '', '2026-10-09 14:29:42.200973 +00:00', '2026-10-09 14:34:26.958291 +00:00', 1);

INSERT INTO sikapj.riwayat_diklat (id, nip, nama_diklat, penyelenggara, jumlah_jp, tanggal_mulai, tanggal_selesai, file_sertifikat_path, file_name, created_at) VALUES (1, '198507152010011012', 'Diklat Pengelolaan Keuangan Daerah', 'BPSDM Provinsi Jawa Timur', 40, '2025-02-10', '2025-02-15', 'uploads/diklat/sertifikat_diklat_budi.pdf', 'sertifikat_diklat_budi.pdf', '2026-10-09 13:53:01.278863 +00:00');
INSERT INTO sikapj.riwayat_diklat (id, nip, nama_diklat, penyelenggara, jumlah_jp, tanggal_mulai, tanggal_selesai, file_sertifikat_path, file_name, created_at) VALUES (2, '198811102012011003', 'Pelatihan Tata Kelola Sistem Pemerintahan Berbasis Elektronik (SPBE)', 'Pusdiklat BSSN', 32, '2025-05-02', '2025-05-06', 'uploads/diklat/sertifikat_spbe_rudi.pdf', 'sertifikat_spbe_rudi.pdf', '2026-10-09 13:53:01.278863 +00:00');

INSERT INTO sikapj.settings (key, value, deskripsi, updated_at) VALUES ('ALERT_N_MONTHS', '3', 'Jangka waktu ambang batas peringatan KGB dan KP sebelum jatuh tempo (dalam bulan)', '2026-10-09 13:52:56.794220 +00:00');

INSERT INTO sikapj.users (id, username, password_hash, role, is_active, last_login, created_at, updated_at) VALUES (4, '198811102012011003', '$2a$12$ebUFlVVY8DYYVHKQ5wqBV.ER5UawuiJM76JOLrslxGhOJEOMPFfZG', 'PEGAWAI', true, null, '2026-10-09 13:52:56.911138 +00:00', '2026-10-09 13:52:56.911138 +00:00');
INSERT INTO sikapj.users (id, username, password_hash, role, is_active, last_login, created_at, updated_at) VALUES (2, '198507152010011012', '$2a$12$ebUFlVVY8DYYVHKQ5wqBV.ER5UawuiJM76JOLrslxGhOJEOMPFfZG', 'PEGAWAI', true, null, '2026-10-09 13:52:56.911138 +00:00', '2026-10-09 13:52:56.911138 +00:00');
INSERT INTO sikapj.users (id, username, password_hash, role, is_active, last_login, created_at, updated_at) VALUES (1, 'admin_umpeg', '$2a$12$ebUFlVVY8DYYVHKQ5wqBV.ER5UawuiJM76JOLrslxGhOJEOMPFfZG', 'ADMIN', true, '2026-10-09 14:35:21.647408 +00:00', '2026-10-09 13:52:56.911138 +00:00', '2026-10-09 13:52:56.911138 +00:00');
INSERT INTO sikapj.users (id, username, password_hash, role, is_active, last_login, created_at, updated_at) VALUES (3, '199003202014022005', '$2a$12$ebUFlVVY8DYYVHKQ5wqBV.ER5UawuiJM76JOLrslxGhOJEOMPFfZG', 'PEGAWAI', true, '2026-10-09 14:43:48.625701 +00:00', '2026-10-09 13:52:56.911138 +00:00', '2026-10-09 13:52:56.911138 +00:00');
