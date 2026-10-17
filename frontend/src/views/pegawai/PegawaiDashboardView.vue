<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import type { Pegawai, EBerkasItem, AlertItem, RiwayatDiklat, ArsipKinerja } from '@/types'
import DokumenViewerModal from '@/components/DokumenViewerModal.vue'
import { MAX_KINERJA_YEAR, validateKinerjaYear, validateContact, validateDateRange, validateFile } from '@/utils/validation'

const profile = ref<Pegawai | null>(null)
const berkasList = ref<EBerkasItem[]>([])
const alerts = ref<AlertItem[]>([])
const diklatList = ref<RiwayatDiklat[]>([])
const kinerjaList = ref<ArsipKinerja[]>([])
const kelengkapan = ref(0)
const activeTab = ref<'berkas' | 'diklat' | 'kinerja'>('berkas')
const isLoading = ref(true)
const loadError = ref('')
const isContactModalOpen = ref(false)
const isSavingContact = ref(false)
const contactError = ref('')
const contactForm = ref({ email: '', no_hp: '' })

// Upload Berkas State
const uploadModalOpen = ref(false)
const selectedDoc = ref<EBerkasItem | null>(null)
const selectedFile = ref<File | null>(null)
const isUploading = ref(false)
const uploadError = ref('')

// Modal Tambah Diklat (FR-08)
const isDiklatModalOpen = ref(false)
const isSavingDiklat = ref(false)
const diklatError = ref('')
const diklatForm = ref({
  nama_diklat: '',
  penyelenggara: '',
  jumlah_jp: 0,
  tanggal_mulai: '',
  tanggal_selesai: '',
})
const diklatFile = ref<File | null>(null)

// Modal Upload Kinerja (FR-07)
const isKinerjaModalOpen = ref(false)
const isSavingKinerja = ref(false)
const kinerjaError = ref('')
const kinerjaForm = ref({
  tahun_periode: new Date().getFullYear(),
  jenis_dokumen: 'SKP',
})
const kinerjaFile = ref<File | null>(null)

// Preview Modal
const previewModalOpen = ref(false)
const previewTitle = ref('')
const previewUrl = ref('')
const previewFileName = ref('')

async function fetchPortalData() {
  isLoading.value = true
  loadError.value = ''
  try {
    const res = await api.get('/pegawai/me/dashboard')
    if (!res.data.success) throw new Error('Data dashboard tidak tersedia')
    profile.value = res.data.data.pegawai
    kelengkapan.value = res.data.data.kelengkapan_berkas
    alerts.value = res.data.data.alerts
    isLoading.value = false

    if (profile.value?.nip) {
      const results = await Promise.allSettled([
        fetchBerkas(profile.value.nip),
        fetchDiklat(profile.value.nip),
        fetchKinerja(profile.value.nip),
      ])
      if (results.some(result => result.status === 'rejected')) {
        loadError.value = 'Sebagian data belum dapat dimuat. Silakan coba lagi.'
      }
    }
  } catch (err) {
    loadError.value = 'Gagal memuat atau memperbarui data. Silakan coba lagi.'
  } finally {
    isLoading.value = false
  }
}

async function fetchBerkas(nip: string) {
  const res = await api.get(`/berkas/${nip}`)
  if (res.data.success) berkasList.value = res.data.data || []
}

async function fetchDiklat(nip: string) {
  const res = await api.get(`/diklat/${nip}`)
  if (res.data.success) diklatList.value = res.data.data || []
}

async function fetchKinerja(nip: string) {
  const res = await api.get(`/kinerja/${nip}`)
  if (res.data.success) kinerjaList.value = res.data.data || []
}

function openContactModal() {
  contactError.value = ''
  contactForm.value = { email: profile.value?.email || '', no_hp: profile.value?.no_hp || '' }
  isContactModalOpen.value = true
}

async function saveContact() {
  contactError.value = validateContact(contactForm.value.email, contactForm.value.no_hp)
  if (contactError.value) return
  isSavingContact.value = true
  try {
    const res = await api.put('/pegawai/me/contact', contactForm.value)
    if (res.data.success) {
      if (profile.value) {
        profile.value.email = contactForm.value.email
        profile.value.no_hp = contactForm.value.no_hp
      }
      isContactModalOpen.value = false
    }
  } catch (err: any) {
    contactError.value = err.response?.data?.message || 'Gagal memperbarui kontak'
  } finally {
    isSavingContact.value = false
  }
}

// Handler Upload Berkas (FR-03)
function openUploadModal(doc: EBerkasItem) {
  selectedDoc.value = doc
  selectedFile.value = null
  uploadError.value = ''
  uploadModalOpen.value = true
}

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) selectedFile.value = target.files[0]
}

async function submitUpload() {
  if (!selectedDoc.value || !selectedFile.value) return
  uploadError.value = validateFile(selectedFile.value, selectedDoc.value.max_size_mb)
  if (uploadError.value) return
  isUploading.value = true

  const formData = new FormData()
  formData.append('dokumen_id', selectedDoc.value.dokumen_id.toString())
  formData.append('file', selectedFile.value)

  try {
    const res = await api.post('/berkas/upload', formData)
    if (res.data.success) {
      uploadModalOpen.value = false
      await fetchPortalData()
    }
  } catch (err: any) {
    uploadError.value = err.response?.data?.message || 'Gagal mengunggah file'
  } finally {
    isUploading.value = false
  }
}

// Handler Diklat (FR-08)
const editingDiklat = ref<RiwayatDiklat | null>(null)
const deletingDiklat = ref<number | null>(null)
function openEditDiklat(d: RiwayatDiklat) {
 openAddDiklatModal()
 editingDiklat.value = d
 diklatForm.value = { nama_diklat: d.nama_diklat, penyelenggara: d.penyelenggara, jumlah_jp: d.jumlah_jp, tanggal_mulai: d.tanggal_mulai.substring(0,10), tanggal_selesai: d.tanggal_selesai.substring(0,10) }
}
async function deleteDiklat(d: RiwayatDiklat) {
 if (!confirm(`Hapus riwayat diklat "${d.nama_diklat}" beserta sertifikatnya?`)) return
 deletingDiklat.value = d.id; loadError.value = ''
 try { await api.delete(`/diklat/${d.id}`); if (profile.value?.nip) await fetchDiklat(profile.value.nip) }
 catch (e: any) { loadError.value = e.response?.data?.message || 'Gagal menghapus diklat.' }
 finally { deletingDiklat.value = null }
}
function openAddDiklatModal() {
  editingDiklat.value = null
  diklatError.value = ''
  diklatForm.value = {
    nama_diklat: '',
    penyelenggara: '',
    jumlah_jp: 30,
    tanggal_mulai: '',
    tanggal_selesai: '',
  }
  diklatFile.value = null
  isDiklatModalOpen.value = true
}

function onDiklatFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) diklatFile.value = target.files[0]
}

async function submitDiklat() {
  diklatError.value = validateDateRange(diklatForm.value.tanggal_mulai, diklatForm.value.tanggal_selesai) || (diklatFile.value ? validateFile(diklatFile.value, 2) : editingDiklat.value?.file_name ? '' : 'Sertifikat wajib diunggah (maks 2 MB).')
  if (diklatError.value) return
  isSavingDiklat.value = true
  const formData = new FormData()
  formData.append('nama_diklat', diklatForm.value.nama_diklat)
  formData.append('penyelenggara', diklatForm.value.penyelenggara)
  formData.append('jumlah_jp', diklatForm.value.jumlah_jp.toString())
  formData.append('tanggal_mulai', diklatForm.value.tanggal_mulai)
  formData.append('tanggal_selesai', diklatForm.value.tanggal_selesai)
  if (diklatFile.value) {
    formData.append('file_sertifikat', diklatFile.value)
  }

  try {
    const res = editingDiklat.value ? await api.put(`/diklat/${editingDiklat.value.id}`, formData) : await api.post('/diklat', formData)
    if (res.data.success) {
      isDiklatModalOpen.value = false
      if (profile.value?.nip) await fetchDiklat(profile.value.nip)
    }
  } catch (err: any) {
    diklatError.value = err.response?.data?.message || 'Gagal menyimpan diklat'
  } finally {
    isSavingDiklat.value = false
  }
}

// Handler Kinerja (FR-07)
function openAddKinerjaModal() {
  kinerjaError.value = ''
  kinerjaForm.value = {
    tahun_periode: new Date().getFullYear(),
    jenis_dokumen: 'SKP',
  }
  kinerjaFile.value = null
  isKinerjaModalOpen.value = true
}

function onKinerjaFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) kinerjaFile.value = target.files[0]
}

async function submitKinerja() {
  if (!kinerjaFile.value) return
  kinerjaError.value = validateKinerjaYear(kinerjaForm.value.tahun_periode) || validateFile(kinerjaFile.value, 1)
  if (kinerjaError.value) return
  isSavingKinerja.value = true

  const formData = new FormData()
  formData.append('tahun_periode', kinerjaForm.value.tahun_periode.toString())
  formData.append('jenis_dokumen', kinerjaForm.value.jenis_dokumen)
  formData.append('file', kinerjaFile.value)

  try {
    const res = await api.post('/kinerja/upload', formData)
    if (res.data.success) {
      isKinerjaModalOpen.value = false
      if (profile.value?.nip) await fetchKinerja(profile.value.nip)
    }
  } catch (err: any) {
    kinerjaError.value = err.response?.data?.message || 'Gagal mengunggah dokumen kinerja'
  } finally {
    isSavingKinerja.value = false
  }
}

// Previews
function openPreview(item: EBerkasItem) {
  if (!item.berkas_id) return
  previewTitle.value = item.nama_dokumen
  previewFileName.value = item.file_name || ''
  previewUrl.value = `/api/berkas/preview/${item.berkas_id}`
  previewModalOpen.value = true
}

function openDiklatPreview(d: RiwayatDiklat) {
  previewTitle.value = `Sertifikat: ${d.nama_diklat}`
  previewFileName.value = d.file_name || ''
  previewUrl.value = `/api/diklat/preview/${d.id}`
  previewModalOpen.value = true
}

function openKinerjaPreview(k: ArsipKinerja) {
  previewTitle.value = `${k.jenis_dokumen} Tahun ${k.tahun_periode}`
  previewFileName.value = k.file_name || ''
  previewUrl.value = `/api/kinerja/preview/${k.id}`
  previewModalOpen.value = true
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(() => {
  fetchPortalData()
})
</script>

<template>
  <div class="space-y-6">
    <p v-if="isLoading" role="status">Memuat data...</p>
    <p v-if="loadError" role="alert" class="p-3 bg-rose-50 text-rose-700 rounded-lg">{{ loadError }}</p>
    <!-- Header Pegawai -->
    <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 rounded-2xl bg-primary-100 text-primary-700 flex items-center justify-center font-bold text-2xl border border-primary-200">
          👤
        </div>
        <div>
          <h1 class="text-xl font-bold text-slate-800">{{ profile?.nama_lengkap }}</h1>
          <p class="text-xs text-slate-500 font-mono">NIP: {{ profile?.nip }} • {{ profile?.pangkat }} ({{ profile?.golongan }})</p>
          <p class="text-xs text-slate-400 mt-0.5">{{ profile?.nama_jabatan }} – {{ profile?.bidang }}</p>
        </div>
      </div>

      <!-- Kartu Ringkasan Cepat -->
      <div class="flex gap-4 items-center">
        <button type="button" @click.stop="openContactModal" class="px-3 py-2 text-xs font-semibold rounded-lg border border-slate-300 hover:bg-slate-50 text-slate-700">
          Edit Kontak
        </button>
        <div class="bg-slate-50 border border-slate-200 rounded-xl px-4 py-2.5 text-center">
          <div class="text-[11px] text-slate-500 uppercase font-semibold">Kelengkapan Berkas</div>
          <div class="text-xl font-bold text-primary-600">{{ kelengkapan.toFixed(0) }}%</div>
        </div>
      </div>
    </div>

    <div v-if="isContactModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <form @submit.prevent="saveContact" class="bg-white rounded-2xl shadow-xl max-w-md w-full p-6 space-y-4">
        <div class="flex items-center justify-between border-b pb-3">
          <h3 class="font-bold text-slate-800">Perbarui Informasi Kontak</h3>
          <button type="button" @click="isContactModalOpen = false" class="text-slate-400">✕</button>
        </div>
        <p v-if="contactError" class="text-xs text-rose-700 bg-rose-50 p-2 rounded">{{ contactError }}</p>
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-1">Email</label>
          <input v-model.trim="contactForm.email" type="email" class="w-full p-2.5 border border-slate-300 rounded-lg" />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-1">Nomor HP/WhatsApp</label>
          <input v-model.trim="contactForm.no_hp" type="tel" class="w-full p-2.5 border border-slate-300 rounded-lg" />
        </div>
        <div class="flex justify-end gap-2">
          <button type="button" @click="isContactModalOpen = false" class="px-4 py-2 bg-slate-100 rounded-lg text-xs font-semibold">Batal</button>
          <button type="submit" :disabled="isSavingContact" class="px-4 py-2 bg-primary-600 text-white rounded-lg text-xs font-semibold disabled:opacity-50">{{ isSavingContact ? 'Menyimpan...' : 'Simpan' }}</button>
        </div>
      </form>
    </div>

    <section v-if="profile" class="bg-white rounded-xl border p-5">
      <h2 class="font-bold mb-3">Identitas &amp; Kepegawaian</h2>
      <dl class="grid sm:grid-cols-2 gap-3 text-sm">
        <div><dt>Tempat / Tanggal Lahir</dt><dd>{{ profile.tempat_lahir || '-' }} / {{ formatDate(profile.tanggal_lahir || '') }}</dd></div>
        <div><dt>Jenis Kelamin</dt><dd>{{ profile.jenis_kelamin === 'L' ? 'Laki-laki' : profile.jenis_kelamin === 'P' ? 'Perempuan' : '-' }}</dd></div>
        <div><dt>Email / No. HP</dt><dd>{{ profile.email || '-' }} / {{ profile.no_hp || '-' }}</dd></div>
        <div><dt>Status / Jenis Jabatan</dt><dd>{{ profile.status_pegawai }} / {{ profile.jenis_jabatan }}</dd></div>
        <div><dt>Pendidikan</dt><dd>{{ profile.jenjang_pendidikan || '-' }} – {{ profile.jurusan || '-' }} – {{ profile.institusi_pendidikan || '-' }}</dd></div>
        <div><dt>TMT Pengangkatan</dt><dd>{{ formatDate(profile.tmt_pengangkatan) }}</dd></div>
        <div><dt>TMT Pangkat Terakhir</dt><dd>{{ formatDate(profile.tmt_pangkat_terakhir) }}</dd></div>
        <div><dt>TMT KGB Terakhir</dt><dd>{{ formatDate(profile.tmt_kgb_terakhir) }}</dd></div>
      </dl>
    </section>
    <!-- Alert Notifikasi Mandiri (KGB/KP) -->
    <div v-for="a in alerts" :key="a.tipe_alert" class="space-y-2">
      <div
        v-if="a.status !== 'AMAN'"
        class="p-4 rounded-xl border flex items-center justify-between"
        :class="a.status === 'SEGERA' ? 'bg-rose-50 border-rose-200 text-rose-900' : 'bg-amber-50 border-amber-200 text-amber-900'"
      >
        <div class="flex items-center gap-3">
          <span class="text-2xl">{{ a.status === 'SEGERA' ? '⚠️' : '🔔' }}</span>
          <div>
            <div class="font-bold text-sm">
              Pemberitahuan {{ a.tipe_alert }} ({{ a.tipe_alert === 'KGB' ? 'Kenaikan Gaji Berkala' : 'Kenaikan Pangkat' }})
            </div>
            <div class="text-xs">
              Jatuh tempo pada <b>{{ formatDate(a.tanggal_jatuh_tempo) }}</b>
              <span v-if="a.sisa_hari < 0"> (Terlambat {{ Math.abs(a.sisa_hari) }} hari)</span>
              <span v-else-if="a.sisa_hari === 0"> (Jatuh tempo hari ini)</span>
              <span v-else-if="a.status === 'SEGERA'"> (Segera dalam {{ a.sisa_hari }} hari lagi)</span>
              <span v-else> (Estimasi {{ a.sisa_bulan.toFixed(1) }} bulan lagi)</span>.
            </div>
          </div>
        </div>
        <span class="text-xs font-semibold px-2.5 py-1 rounded-full uppercase" :class="a.status === 'SEGERA' ? 'bg-rose-200 text-rose-800' : 'bg-amber-200 text-amber-800'">
          {{ a.status }}
        </span>
      </div>
    </div>

    <!-- Tabs Navigasi Portal Mandiri -->
    <div class="flex border-b border-slate-200 gap-4">
      <button
        type="button"
        @click="activeTab = 'berkas'"
        class="pb-3 px-3 text-sm font-semibold border-b-2 transition-all flex items-center gap-2 cursor-pointer outline-none"
        :class="activeTab === 'berkas' ? 'border-primary-600 text-primary-700' : 'border-transparent text-slate-500 hover:text-slate-700'"
      >
        <span>📁 12 Dokumen E-Berkas</span>
        <span class="text-xs px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 font-mono">{{ berkasList.length }}</span>
      </button>

      <button
        type="button"
        @click="activeTab = 'diklat'"
        class="pb-3 px-3 text-sm font-semibold border-b-2 transition-all flex items-center gap-2 cursor-pointer outline-none"
        :class="activeTab === 'diklat' ? 'border-primary-600 text-primary-700' : 'border-transparent text-slate-500 hover:text-slate-700'"
      >
        <span>🎓 Rekam Jejak Diklat</span>
        <span class="text-xs px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 font-mono">{{ diklatList.length }}</span>
      </button>

      <button
        type="button"
        @click="activeTab = 'kinerja'"
        class="pb-3 px-3 text-sm font-semibold border-b-2 transition-all flex items-center gap-2 cursor-pointer outline-none"
        :class="activeTab === 'kinerja' ? 'border-primary-600 text-primary-700' : 'border-transparent text-slate-500 hover:text-slate-700'"
      >
        <span>📊 Arsip Kinerja SKP/PK</span>
        <span class="text-xs px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 font-mono">{{ kinerjaList.length }}</span>
      </button>
    </div>

    <!-- TAB 1: 12 DOKUMEN E-BERKAS -->
    <div v-if="activeTab === 'berkas'" class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6 space-y-4">
      <div class="flex items-center justify-between border-b border-slate-100 pb-3">
        <div>
          <h2 class="text-lg font-bold text-slate-800">Daftar Dokumen E-Berkas (Digital Dossier)</h2>
          <p class="text-xs text-slate-500">Unggah berkas persyaratan (Format PDF/JPG/PNG).</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5">
        <div
          v-for="doc in berkasList"
          :key="doc.dokumen_id"
          class="p-4 rounded-xl border border-slate-200 hover:border-slate-300 transition-all flex flex-col justify-between space-y-3"
        >
          <div>
            <div class="flex items-start justify-between gap-2">
              <h4 class="font-bold text-sm text-slate-800">{{ doc.nama_dokumen }}</h4>
              <span
                class="text-[11px] font-semibold px-2 py-0.5 rounded-full whitespace-nowrap"
                :class="{
                  'bg-emerald-100 text-emerald-800': doc.status_verifikasi === 'Disetujui',
                  'bg-amber-100 text-amber-800': doc.status_verifikasi === 'Menunggu Verifikasi',
                  'bg-rose-100 text-rose-800': doc.status_verifikasi === 'Ditolak',
                  'bg-slate-100 text-slate-500': doc.status_verifikasi === 'Belum Diunggah'
                }"
              >
                {{ doc.status_verifikasi }}
              </span>
            </div>
            <p class="text-xs text-slate-400 mt-1">Batas ukuran: maks {{ doc.max_size_mb }} MB</p>

            <div v-if="doc.file_name" class="text-xs text-slate-600 mt-1.5 font-mono truncate">
              📄 {{ doc.file_name }}
            </div>
            <div v-if="doc.catatan_admin" class="mt-2 text-xs bg-rose-50 text-rose-700 p-2 rounded-lg border border-rose-200">
              <b>Catatan Revisi:</b> {{ doc.catatan_admin }}
            </div>
          </div>

          <!-- Tombol Aksi -->
          <div class="flex items-center gap-2 pt-2 border-t border-slate-100">
            <button
              v-if="doc.berkas_id"
              @click="openPreview(doc)"
              class="flex-1 py-1.5 px-3 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-semibold transition-colors flex items-center justify-center gap-1"
            >
              👁️ Lihat Berkas
            </button>
            <button type="button" @click.stop="openUploadModal(doc)"
              class="flex-1 py-1.5 px-3 bg-primary-600 hover:bg-primary-700 text-white rounded-lg text-xs font-semibold transition-colors text-center"
            >
              {{ doc.berkas_id ? 'Unggah Ulang' : 'Unggah File' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 2: REKAM JEJAK DIKLAT & SERTIFIKAT (FR-08) -->
    <div v-if="activeTab === 'diklat'" class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6 space-y-4">
      <div class="flex items-center justify-between border-b border-slate-100 pb-3">
        <div>
          <h2 class="text-lg font-bold text-slate-800">Rekam Jejak Pengembangan Kompetensi & Diklat</h2>
          <p class="text-xs text-slate-500">Catat riwayat pelatihan teknis/fungsional beserta unggah sertifikat bukti fisik.</p>
        </div>
        <button
          @click="openAddDiklatModal"
          class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-xs font-semibold rounded-xl flex items-center gap-1.5 shadow-sm"
        >
          <span>➕ Tambah Diklat</span>
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-slate-600 text-xs font-semibold uppercase border-b border-slate-200">
            <tr>
              <th class="px-5 py-3">Nama Diklat / Pelatihan</th>
              <th class="px-5 py-3">Penyelenggara</th>
              <th class="px-5 py-3 text-center">Durasi (JP)</th>
              <th class="px-5 py-3">Periode Pelaksanaan</th>
              <th class="px-5 py-3 text-right">Sertifikat</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700">
            <tr v-for="d in diklatList" :key="d.id" class="hover:bg-slate-50/80">
              <td class="px-5 py-3.5 font-bold text-slate-800">{{ d.nama_diklat }}</td>
              <td class="px-5 py-3.5 text-xs text-slate-600">{{ d.penyelenggara }}</td>
              <td class="px-5 py-3.5 text-center text-xs font-semibold">{{ d.jumlah_jp }} JP</td>
              <td class="px-5 py-3.5 text-xs font-mono">
                {{ formatDate(d.tanggal_mulai) }} s.d. {{ formatDate(d.tanggal_selesai) }}
              </td>
              <td class="px-5 py-3.5 text-right">
                <button @click="openEditDiklat(d)" class="mr-2 text-primary-700">Edit</button><button @click="deleteDiklat(d)" :disabled="deletingDiklat !== null" class="mr-2 text-rose-700">Hapus</button>
                <button
                  v-if="d.file_name"
                  @click="openDiklatPreview(d)"
                  class="px-3 py-1 bg-primary-50 hover:bg-primary-100 text-primary-700 border border-primary-200 rounded-lg text-xs font-semibold inline-flex items-center gap-1"
                >
                  👁️ Lihat Sertifikat
                </button>
                <span v-else class="text-xs text-slate-400 italic">Tidak ada file</span>
              </td>
            </tr>
            <tr v-if="diklatList.length === 0">
              <td colspan="5" class="px-6 py-8 text-center text-slate-400 text-xs">
                Belum ada rekaman diklat yang ditambahkan.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- TAB 3: ARSIP KINERJA (FR-07) -->
    <div v-if="activeTab === 'kinerja'" class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6 space-y-4">
      <div class="flex items-center justify-between border-b border-slate-100 pb-3">
        <div>
          <h2 class="text-lg font-bold text-slate-800">Arsip Kinerja (Perjanjian Kinerja & SKP)</h2>
          <p class="text-xs text-slate-500">Pengarsipan dokumen evaluasi kinerja per tahun periode.</p>
        </div>
        <button
          @click="openAddKinerjaModal"
          class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-xs font-semibold rounded-xl flex items-center gap-1.5 shadow-sm"
        >
          <span>➕ Unggah Dokumen Kinerja</span>
        </button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div
          v-for="k in kinerjaList"
          :key="k.id"
          class="p-4 border rounded-xl hover:border-slate-300 transition-all flex flex-col justify-between space-y-3"
        >
          <div>
            <div class="flex justify-between items-center">
              <span class="px-2.5 py-0.5 rounded font-bold text-xs" :class="k.jenis_dokumen === 'SKP' ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'">
                {{ k.jenis_dokumen }}
              </span>
              <span class="text-sm font-bold text-slate-800 font-mono">{{ k.tahun_periode }}</span>
            </div>
            <p class="text-xs text-slate-500 mt-2 font-mono truncate">📄 {{ k.file_name }}</p>
          </div>
          <button
            @click="openKinerjaPreview(k)"
            class="w-full py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-semibold"
          >
            👁️ Lihat Dokumen
          </button>
        </div>
        <div v-if="kinerjaList.length === 0" class="col-span-3 text-center py-8 text-slate-400 text-xs">
          Belum ada arsip dokumen kinerja tahunan yang diunggah.
        </div>
      </div>
    </div>

    <!-- MODAL UNGGAH BERKAS -->
    <div v-if="uploadModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="bg-white rounded-xl shadow-xl max-w-md w-full p-6 space-y-4">
        <h3 class="text-base font-bold text-slate-800">Unggah {{ selectedDoc?.nama_dokumen }}</h3>
        <p class="text-xs text-slate-500">Maksimal ukuran {{ selectedDoc?.max_size_mb }} MB (PDF/JPG/PNG)</p>
        <div v-if="uploadError" class="p-2.5 rounded bg-rose-50 text-rose-700 text-xs">{{ uploadError }}</div>
        <div class="border-2 border-dashed border-slate-200 rounded-xl p-6 text-center space-y-2">
          <input type="file" accept=".pdf,.jpg,.jpeg,.png" @change="onFileChange" class="text-xs text-slate-500" />
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button @click="uploadModalOpen = false" class="px-4 py-2 bg-slate-100 text-slate-700 text-xs font-semibold rounded-lg">Batal</button>
          <button @click="submitUpload" :disabled="!selectedFile || isUploading" class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:opacity-50 text-white text-xs font-semibold rounded-lg">
            {{ isUploading ? 'Mengunggah...' : 'Simpan Berkas' }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL TAMBAH DIKLAT (FR-08) -->
    <div v-if="isDiklatModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="bg-white rounded-xl shadow-xl max-w-lg w-full p-6 space-y-4">
        <h3 class="text-base font-bold text-slate-800">{{ editingDiklat ? 'Edit Rekam Jejak Diklat' : 'Tambah Rekam Jejak Diklat' }}</h3>
        <div v-if="diklatError" class="p-2.5 rounded bg-rose-50 text-rose-700 text-xs">{{ diklatError }}</div>
        <form @submit.prevent="submitDiklat" class="space-y-3 text-xs">
          <div>
            <label class="block font-semibold text-slate-600 mb-1">Nama Diklat / Pelatihan *</label>
            <input v-model="diklatForm.nama_diklat" type="text" required placeholder="Contoh: Diklat Pengelolaan Keuangan Daerah" class="w-full p-2 border rounded-lg" />
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">Penyelenggara *</label>
            <input v-model="diklatForm.penyelenggara" type="text" required placeholder="Contoh: BPSDM Provinsi Jawa Timur" class="w-full p-2 border rounded-lg" />
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">Jumlah Jam Pelajaran (JP) *</label>
            <input v-model.number="diklatForm.jumlah_jp" type="number" required min="1" class="w-full p-2 border rounded-lg" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Tanggal Mulai *</label>
              <input v-model="diklatForm.tanggal_mulai" type="date" required class="w-full p-2 border rounded-lg" />
            </div>
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Tanggal Selesai *</label>
              <input v-model="diklatForm.tanggal_selesai" type="date" required class="w-full p-2 border rounded-lg" />
            </div>
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">File Sertifikat (PDF / JPG / PNG, maks 2 MB)</label>
            <p v-if="editingDiklat?.file_name" class="text-slate-500 mb-2">Sertifikat saat ini: {{ editingDiklat.file_name }}. Kosongkan untuk mempertahankan.</p>
            <input type="file" accept=".pdf,.jpg,.jpeg,.png" :required="!editingDiklat?.file_name" @change="onDiklatFileChange" class="w-full p-2 border rounded-lg text-slate-500" />
          </div>
          <div class="flex justify-end gap-2 pt-2 border-t">
            <button type="button" @click="isDiklatModalOpen = false" class="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg font-semibold">Batal</button>
            <button type="submit" :disabled="isSavingDiklat" class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-semibold">
              {{ isSavingDiklat ? 'Menyimpan...' : 'Simpan Diklat' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL UNGGAH KINERJA (FR-07) -->
    <div v-if="isKinerjaModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="bg-white rounded-xl shadow-xl max-w-md w-full p-6 space-y-4">
        <h3 class="text-base font-bold text-slate-800">Unggah Dokumen Kinerja (PK / SKP)</h3>
        <div v-if="kinerjaError" class="p-2.5 rounded bg-rose-50 text-rose-700 text-xs">{{ kinerjaError }}</div>
        <form @submit.prevent="submitKinerja" class="space-y-3 text-xs">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Jenis Dokumen *</label>
              <select v-model="kinerjaForm.jenis_dokumen" class="w-full p-2 border rounded-lg">
                <option value="SKP">SKP</option>
                <option value="PK">Perjanjian Kinerja (PK)</option>
              </select>
            </div>
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Tahun Periode *</label>
              <input v-model.number="kinerjaForm.tahun_periode" type="number" required min="2000" :max="MAX_KINERJA_YEAR" class="w-full p-2 border rounded-lg font-mono" />
            </div>
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">File Dokumen PDF / JPG / PNG (Maks 1 MB) *</label>
            <input type="file" accept=".pdf,.jpg,.jpeg,.png" required @change="onKinerjaFileChange" class="w-full p-2 border rounded-lg text-slate-500" />
          </div>
          <div class="flex justify-end gap-2 pt-2 border-t">
            <button type="button" @click="isKinerjaModalOpen = false" class="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg font-semibold">Batal</button>
            <button type="submit" :disabled="isSavingKinerja || !kinerjaFile" class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-semibold">
              {{ isSavingKinerja ? 'Mengunggah...' : 'Simpan Kinerja' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- PREVIEW MODAL -->
    <DokumenViewerModal
      :is-open="previewModalOpen"
      :title="previewTitle"
      :file-name="previewFileName"
      :file-url="previewUrl"
      @close="previewModalOpen = false"
    />
  </div>
</template>
