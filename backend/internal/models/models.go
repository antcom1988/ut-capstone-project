package models

import "time"

// User merepresentasikan akun autentikasi
type User struct {
	ID           int        `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         string     `db:"role" json:"role"` // ADMIN / PEGAWAI
	IsActive     bool       `db:"is_active" json:"is_active"`
	LastLogin    *time.Time `db:"last_login" json:"last_login"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// Pegawai merepresentasikan master data detail pegawai
type Pegawai struct {
	NIP                 string    `db:"nip" json:"nip"`
	UserID              *int      `db:"user_id" json:"user_id"`
	NamaLengkap         string    `db:"nama_lengkap" json:"nama_lengkap"`
	TempatLahir         *string   `db:"tempat_lahir" json:"tempat_lahir"`
	TanggalLahir        *time.Time `db:"tanggal_lahir" json:"tanggal_lahir"`
	JenisKelamin        *string   `db:"jenis_kelamin" json:"jenis_kelamin"`
	Email               *string   `db:"email" json:"email"`
	NoHP                *string   `db:"no_hp" json:"no_hp"`
	StatusPegawai       string    `db:"status_pegawai" json:"status_pegawai"` // PNS / PPPK
	Bidang              string    `db:"bidang" json:"bidang"`
	JenisJabatan        string    `db:"jenis_jabatan" json:"jenis_jabatan"`
	NamaJabatan         string    `db:"nama_jabatan" json:"nama_jabatan"`
	Pangkat             string    `db:"pangkat" json:"pangkat"`
	Golongan            string    `db:"golongan" json:"golongan"`
	TMTPengangkatan     time.Time `db:"tmt_pengangkatan" json:"tmt_pengangkatan"`
	TMTPangkatTerakhir  time.Time `db:"tmt_pangkat_terakhir" json:"tmt_pangkat_terakhir"`
	TMTKGBTerakhir      time.Time `db:"tmt_kgb_terakhir" json:"tmt_kgb_terakhir"`
	JenjangPendidikan   *string   `db:"jenjang_pendidikan" json:"jenjang_pendidikan"`
	Jurusan             *string   `db:"jurusan" json:"jurusan"`
	InstitusiPendidikan *string   `db:"institusi_pendidikan" json:"institusi_pendidikan"`
	IsActive            bool      `db:"is_active" json:"is_active"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// MasterDokumen merepresentasikan daftar 12 dokumen persyaratan
type MasterDokumen struct {
	ID          int     `db:"id" json:"id"`
	NamaDokumen string  `db:"nama_dokumen" json:"nama_dokumen"`
	MaxSizeMB   int     `db:"max_size_mb" json:"max_size_mb"`
	IsRequired  bool    `db:"is_required" json:"is_required"`
	Deskripsi   *string `db:"deskripsi" json:"deskripsi"`
}

// EBerkas merepresentasikan dokumen digital yang diunggah pegawai
type EBerkas struct {
	ID               int        `db:"id" json:"id"`
	NIP              string     `db:"nip" json:"nip"`
	DokumenID        int        `db:"dokumen_id" json:"dokumen_id"`
	NamaDokumen      string     `db:"nama_dokumen" json:"nama_dokumen"`
	MaxSizeMB        int        `db:"max_size_mb" json:"max_size_mb"`
	FilePath         string     `db:"file_path" json:"file_path"`
	FileName         string     `db:"file_name" json:"file_name"`
	FileSizeBytes    int64      `db:"file_size_bytes" json:"file_size_bytes"`
	MimeType         string     `db:"mime_type" json:"mime_type"`
	StatusVerifikasi string     `db:"status_verifikasi" json:"status_verifikasi"` // Menunggu Verifikasi / Disetujui / Ditolak
	CatatanAdmin     *string    `db:"catatan_admin" json:"catatan_admin"`
	UploadedAt       time.Time  `db:"uploaded_at" json:"uploaded_at"`
	VerifiedAt       *time.Time `db:"verified_at" json:"verified_at"`
	VerifiedBy       *int       `db:"verified_by" json:"verified_by"`
}

// ArsipKinerja merepresentasikan dokumen tahunan PK dan SKP
type ArsipKinerja struct {
	ID            int       `db:"id" json:"id"`
	NIP           string    `db:"nip" json:"nip"`
	TahunPeriode  int       `db:"tahun_periode" json:"tahun_periode"`
	JenisDokumen  string    `db:"jenis_dokumen" json:"jenis_dokumen"` // PK / SKP
	FilePath      string    `db:"file_path" json:"file_path"`
	FileName      string    `db:"file_name" json:"file_name"`
	FileSizeBytes int64     `db:"file_size_bytes" json:"file_size_bytes"`
	UploadedAt    time.Time `db:"uploaded_at" json:"uploaded_at"`
}

// RiwayatDiklat merepresentasikan rekaman pelatihan/kompetensi
type RiwayatDiklat struct {
	ID                 int        `db:"id" json:"id"`
	NIP                string     `db:"nip" json:"nip"`
	NamaDiklat         string     `db:"nama_diklat" json:"nama_diklat"`
	Penyelenggara      string     `db:"penyelenggara" json:"penyelenggara"`
	JumlahJP           int        `db:"jumlah_jp" json:"jumlah_jp"`
	TanggalMulai       time.Time  `db:"tanggal_mulai" json:"tanggal_mulai"`
	TanggalSelesai     time.Time  `db:"tanggal_selesai" json:"tanggal_selesai"`
	FileSertifikatPath *string    `db:"file_sertifikat_path" json:"file_sertifikat_path"`
	FileName           *string    `db:"file_name" json:"file_name"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
}

// AlertItem merepresentasikan proyeksi jatuh tempo KGB / KP
type AlertItem struct {
	NIP              string    `json:"nip"`
	NamaLengkap      string    `json:"nama_lengkap"`
	Bidang           string    `json:"bidang"`
	PangkatGolongan  string    `json:"pangkat_golongan"`
	TipeAlert        string    `json:"tipe_alert"` // "KGB" atau "KP"
	TMTDasar         time.Time `json:"tmt_dasar"`
	TanggalJatuhTempo time.Time `json:"tanggal_jatuh_tempo"`
	SisaBulan        float64   `json:"sisa_bulan"`
	SisaHari         int       `json:"sisa_hari"`
	Status           string    `json:"status"` // "SEGERA", "DIINGATKAN", "AMAN"
}

// DashboardSummary merepresentasikan metrik agregat di dashboard
type DashboardSummary struct {
	TotalPegawai        int         `json:"total_pegawai"`
	BerkasLengkapPersen float64     `json:"berkas_lengkap_persen"`
	MenungguVerifikasi  int         `json:"menunggu_verifikasi"`
	AlertMendekati      int         `json:"alert_mendekati"`
	Alerts              []AlertItem `json:"alerts"`
}
