export interface User {
  id: number
  username: string
  role: 'ADMIN' | 'PEGAWAI'
  nip?: string
  must_change_password?: boolean
}

export interface Pegawai {
  nip: string
  user_id?: number
  nama_lengkap: string
  tempat_lahir?: string
  tanggal_lahir?: string
  jenis_kelamin?: 'L' | 'P'
  email?: string
  no_hp?: string
  status_pegawai: 'PNS' | 'PPPK'
  bidang: string
  jenis_jabatan: 'Struktural' | 'Fungsional' | 'Pelaksana'
  nama_jabatan: string
  pangkat: string
  golongan: string
  tmt_pengangkatan: string
  tmt_pangkat_terakhir: string
  tmt_kgb_terakhir: string
  jenjang_pendidikan?: string
  jurusan?: string
  institusi_pendidikan?: string
  is_active: boolean
}

export interface MasterDokumen {
  id: number
  nama_dokumen: string
  max_size_mb: number
  is_required: boolean
  deskripsi?: string
}

export interface EBerkasItem {
  dokumen_id: number
  nama_dokumen: string
  max_size_mb: number
  is_required: boolean
  berkas_id?: number
  file_name?: string
  file_size_bytes?: number
  mime_type?: string
  status_verifikasi: 'Belum Diunggah' | 'Menunggu Verifikasi' | 'Disetujui' | 'Ditolak'
  catatan_admin?: string
  uploaded_at?: string
  verified_at?: string
}

export interface AlertItem {
  nip: string
  nama_lengkap: string
  bidang: string
  pangkat_golongan: string
  tipe_alert: 'KGB' | 'KP'
  tmt_dasar: string
  tanggal_jatuh_tempo: string
  sisa_bulan: number
  sisa_hari: number
  status: 'SEGERA' | 'DIINGATKAN' | 'AMAN'
}

export interface DashboardSummary {
  total_pegawai: number
  berkas_lengkap_persen: number
  menunggu_verifikasi: number
  alert_mendekati: number
  alerts: AlertItem[]
  threshold_months?: number
}

export interface ArsipKinerja {
  id: number
  nip: string
  tahun_periode: number
  jenis_dokumen: 'PK' | 'SKP'
  file_name: string
  uploaded_at: string
}

export interface AuditLog {
  id: number
  username?: string
  action: string
  details?: string
  ip_address?: string
  created_at: string
}

export interface RiwayatDiklat {
  id: number
  nip: string
  nama_diklat: string
  penyelenggara: string
  jumlah_jp: number
  tanggal_mulai: string
  tanggal_selesai: string
  file_name?: string
}
