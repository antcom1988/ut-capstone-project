<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import type { Pegawai, EBerkasItem, RiwayatDiklat, ArsipKinerja } from '@/types'
import DokumenViewerModal from '@/components/DokumenViewerModal.vue'

const pegawais = ref<Pegawai[]>([])
const selectedPegawai = ref<Pegawai | null>(null)
const berkasList = ref<EBerkasItem[]>([])
const diklatList = ref<RiwayatDiklat[]>([])
const kinerjaList = ref<ArsipKinerja[]>([])
const adminTab = ref<'berkas' | 'diklat' | 'kinerja'>('berkas')
const isLoading = ref(true)

// Modal Preview Dokumen
const previewModalOpen = ref(false)
const previewTitle = ref('')
const previewUrl = ref('')
const previewFileName = ref('')

// Form Verifikasi
const verifyTarget = ref<EBerkasItem | null>(null)
const verifyStatus = ref<'Disetujui' | 'Ditolak'>('Disetujui')
const verifyCatatan = ref('')
const isSubmittingVerify = ref(false)

async function fetchPegawaiList() {
  isLoading.value = true
  try {
    const res = await api.get('/admin/pegawai')
    if (res.data.success) {
      pegawais.value = res.data.data
      if (pegawais.value.length > 0 && !selectedPegawai.value) {
        selectPegawai(pegawais.value[0])
      }
    }
  } catch (err) {
    console.error(err)
  } finally {
    isLoading.value = false
  }
}

async function selectPegawai(p: Pegawai) {
  selectedPegawai.value = p
  try {
    const [bRes, dRes, kRes] = await Promise.all([
      api.get(`/berkas/${p.nip}`),
      api.get(`/diklat/${p.nip}`),
      api.get(`/kinerja/${p.nip}`),
    ])
    if (bRes.data.success) berkasList.value = bRes.data.data
    if (dRes.data.success) diklatList.value = dRes.data.data
    if (kRes.data.success) kinerjaList.value = kRes.data.data
  } catch (err) {
    console.error(err)
  }
}

function openPreview(item: EBerkasItem) {
  if (!item.berkas_id) return
  const token = localStorage.getItem('sikapj_token') || ''
  previewTitle.value = `${item.nama_dokumen} – ${selectedPegawai.value?.nama_lengkap}`
  previewFileName.value = item.file_name || ''
  previewUrl.value = `/api/berkas/preview/${item.berkas_id}?token=${token}`
  previewModalOpen.value = true
}

function openDiklatPreview(d: RiwayatDiklat) {
  const token = localStorage.getItem('sikapj_token') || ''
  previewTitle.value = `Sertifikat Diklat: ${d.nama_diklat} – ${selectedPegawai.value?.nama_lengkap}`
  previewFileName.value = d.file_name || ''
  previewUrl.value = `/api/diklat/preview/${d.id}?token=${token}`
  previewModalOpen.value = true
}

function openKinerjaPreview(k: ArsipKinerja) {
  const token = localStorage.getItem('sikapj_token') || ''
  previewTitle.value = `${k.jenis_dokumen} ${k.tahun_periode} – ${selectedPegawai.value?.nama_lengkap}`
  previewFileName.value = k.file_name || ''
  previewUrl.value = `/api/kinerja/preview/${k.id}?token=${token}`
  previewModalOpen.value = true
}

function openVerifyModal(item: EBerkasItem) {
  verifyTarget.value = item
  verifyStatus.value = 'Disetujui'
  verifyCatatan.value = ''
}

async function submitVerifikasi() {
  if (!verifyTarget.value?.berkas_id) return
  isSubmittingVerify.value = true
  try {
    const res = await api.put(`/admin/berkas/verify/${verifyTarget.value.berkas_id}`, {
      status_verifikasi: verifyStatus.value,
      catatan_admin: verifyCatatan.value,
    })
    if (res.data.success && selectedPegawai.value) {
      verifyTarget.value = null
      await selectPegawai(selectedPegawai.value)
    }
  } catch (err) {
    console.error(err)
  } finally {
    isSubmittingVerify.value = false
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(() => {
  fetchPegawaiList()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-slate-800">Manajemen E-Berkas & Kinerja Pegawai</h1>
      <p class="text-sm text-slate-500">Verifikasi 12 dokumen e-berkas, pantau rekam jejak diklat, dan arsip SKP/PK per pegawai</p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Kolom Kiri: Daftar Pegawai -->
      <div class="lg:col-span-4 bg-white rounded-xl border border-slate-200 shadow-sm p-4 h-[75vh] flex flex-col">
        <h3 class="text-sm font-semibold text-slate-700 uppercase tracking-wider mb-3 px-2">Daftar Pegawai</h3>
        <div class="overflow-y-auto space-y-1.5 flex-1 pr-1">
          <button
            v-for="p in pegawais"
            :key="p.nip"
            @click="selectPegawai(p)"
            class="w-full text-left p-3 rounded-lg border transition-all text-sm"
            :class="selectedPegawai?.nip === p.nip ? 'border-primary-500 bg-primary-50/70 text-primary-900 font-medium' : 'border-slate-100 hover:bg-slate-50 text-slate-700'"
          >
            <div class="font-medium truncate">{{ p.nama_lengkap }}</div>
            <div class="text-xs text-slate-400 font-mono">{{ p.nip }}</div>
            <div class="text-xs text-slate-500 mt-1 flex justify-between">
              <span>{{ p.bidang }}</span>
              <span class="font-semibold">{{ p.golongan }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Kolom Kanan: Detail Dokumen Pegawai Terpilih -->
      <div class="lg:col-span-8 bg-white rounded-xl border border-slate-200 shadow-sm p-6 h-[75vh] flex flex-col">
        <div v-if="selectedPegawai" class="flex-1 flex flex-col overflow-hidden">
          <div class="border-b border-slate-100 pb-3 mb-3 flex items-center justify-between">
            <div>
              <h2 class="text-lg font-bold text-slate-800">{{ selectedPegawai.nama_lengkap }}</h2>
              <p class="text-xs text-slate-500">{{ selectedPegawai.nama_jabatan }} • NIP: {{ selectedPegawai.nip }}</p>
            </div>
          </div>

          <!-- Nav Tabs Admin -->
          <div class="flex border-b border-slate-100 mb-3 gap-3 text-xs">
            <button
              type="button"
              @click="adminTab = 'berkas'"
              class="pb-2 font-bold border-b-2 transition-all flex items-center gap-1.5 cursor-pointer outline-none"
              :class="adminTab === 'berkas' ? 'border-primary-600 text-primary-700' : 'border-transparent text-slate-400 hover:text-slate-600'"
            >
              <span>📁 12 E-Berkas</span>
            </button>
            <button
              type="button"
              @click="adminTab = 'diklat'"
              class="pb-2 font-bold border-b-2 transition-all flex items-center gap-1.5 cursor-pointer outline-none"
              :class="adminTab === 'diklat' ? 'border-primary-600 text-primary-700' : 'border-transparent text-slate-400 hover:text-slate-600'"
            >
              <span>🎓 Rekam Diklat</span>
              <span class="px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-500 font-mono font-normal">{{ diklatList.length }}</span>
            </button>
            <button
              type="button"
              @click="adminTab = 'kinerja'"
              class="pb-2 font-bold border-b-2 transition-all flex items-center gap-1.5 cursor-pointer outline-none"
              :class="adminTab === 'kinerja' ? 'border-primary-600 text-primary-700' : 'border-transparent text-slate-400 hover:text-slate-600'"
            >
              <span>📊 Arsip SKP/PK</span>
              <span class="px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-500 font-mono font-normal">{{ kinerjaList.length }}</span>
            </button>
          </div>

          <!-- KONTEN TAB 1: 12 BERKAS -->
          <div v-if="adminTab === 'berkas'" class="overflow-y-auto flex-1 space-y-2.5 pr-2">
            <div
              v-for="item in berkasList"
              :key="item.dokumen_id"
              class="p-3.5 rounded-lg border border-slate-200 flex items-center justify-between hover:border-slate-300 transition-colors"
            >
              <div class="space-y-0.5">
                <div class="font-semibold text-sm text-slate-800 flex items-center gap-2">
                  <span>{{ item.nama_dokumen }}</span>
                  <span class="text-[11px] font-normal text-slate-400">maks {{ item.max_size_mb }} MB</span>
                </div>
                <div class="text-xs text-slate-500" v-if="item.file_name">
                  File: <span class="font-mono text-slate-700">{{ item.file_name }}</span>
                </div>
                <div v-if="item.catatan_admin" class="text-xs text-rose-600 bg-rose-50 px-2 py-0.5 rounded inline-block">
                  Catatan: {{ item.catatan_admin }}
                </div>
              </div>

              <!-- Status & Aksi -->
              <div class="flex items-center gap-2">
                <span
                  class="text-xs font-semibold px-2.5 py-1 rounded-full"
                  :class="{
                    'bg-emerald-100 text-emerald-800': item.status_verifikasi === 'Disetujui',
                    'bg-amber-100 text-amber-800': item.status_verifikasi === 'Menunggu Verifikasi',
                    'bg-rose-100 text-rose-800': item.status_verifikasi === 'Ditolak',
                    'bg-slate-100 text-slate-500': item.status_verifikasi === 'Belum Diunggah'
                  }"
                >
                  {{ item.status_verifikasi }}
                </span>

                <button
                  v-if="item.berkas_id"
                  @click="openPreview(item)"
                  class="px-3 py-1 bg-primary-50 hover:bg-primary-100 text-primary-700 rounded-lg text-xs font-semibold transition-colors flex items-center gap-1 border border-primary-200"
                >
                  <span>👁️ Lihat</span>
                </button>

                <button
                  v-if="item.berkas_id && item.status_verifikasi === 'Menunggu Verifikasi'"
                  @click="openVerifyModal(item)"
                  class="px-3 py-1 bg-slate-800 hover:bg-slate-900 text-white rounded-lg text-xs font-semibold transition-colors shadow-sm"
                >
                  Verifikasi
                </button>
              </div>
            </div>
          </div>

          <!-- KONTEN TAB 2: REKAM JEJAK DIKLAT (FR-08) -->
          <div v-if="adminTab === 'diklat'" class="overflow-y-auto flex-1 pr-2 space-y-3">
            <div v-for="d in diklatList" :key="d.id" class="p-3.5 border rounded-lg hover:border-slate-300 flex justify-between items-center text-xs">
              <div>
                <div class="font-bold text-slate-800 text-sm">{{ d.nama_diklat }}</div>
                <div class="text-slate-500 mt-0.5">Penyelenggara: <b>{{ d.penyelenggara }}</b> • Durasi: <b>{{ d.jumlah_jp }} JP</b></div>
                <div class="text-slate-400 font-mono mt-0.5">{{ formatDate(d.tanggal_mulai) }} s.d. {{ formatDate(d.tanggal_selesai) }}</div>
              </div>
              <button
                v-if="d.file_name"
                @click="openDiklatPreview(d)"
                class="px-3 py-1.5 bg-primary-50 hover:bg-primary-100 text-primary-700 font-semibold rounded-lg border border-primary-200"
              >
                👁️ Lihat Sertifikat
              </button>
              <span v-else class="text-slate-400 italic">Tanpa berkas</span>
            </div>
            <div v-if="diklatList.length === 0" class="text-center py-10 text-slate-400 text-xs">
              Pegawai ini belum memiliki rekaman diklat/pelatihan.
            </div>
          </div>

          <!-- KONTEN TAB 3: ARSIP KINERJA (FR-07) -->
          <div v-if="adminTab === 'kinerja'" class="overflow-y-auto flex-1 pr-2 space-y-3">
            <div v-for="k in kinerjaList" :key="k.id" class="p-3.5 border rounded-lg hover:border-slate-300 flex justify-between items-center text-xs">
              <div>
                <div class="font-bold text-slate-800 text-sm flex items-center gap-2">
                  <span>{{ k.jenis_dokumen }} Tahun {{ k.tahun_periode }}</span>
                </div>
                <div class="text-slate-400 font-mono mt-0.5 truncate max-w-md">📄 {{ k.file_name }}</div>
              </div>
              <button
                @click="openKinerjaPreview(k)"
                class="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold rounded-lg"
              >
                👁️ Lihat Dokumen
              </button>
            </div>
            <div v-if="kinerjaList.length === 0" class="text-center py-10 text-slate-400 text-xs">
              Belum ada arsip SKP/PK untuk pegawai ini.
            </div>
          </div>
        </div>

        <div v-else class="flex-1 flex items-center justify-center text-slate-400 text-sm">
          Pilih pegawai di kolom kiri untuk melihat berkas.
        </div>
      </div>
    </div>

    <!-- Modal Verifikasi Admin -->
    <div v-if="verifyTarget" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="bg-white rounded-xl shadow-xl max-w-md w-full p-6 space-y-4">
        <h3 class="text-base font-bold text-slate-800">Verifikasi {{ verifyTarget.nama_dokumen }}</h3>
        <p class="text-xs text-slate-500">Pegawai: {{ selectedPegawai?.nama_lengkap }}</p>

        <div>
          <label class="block text-xs font-semibold text-slate-600 mb-1.5">Keputusan</label>
          <div class="grid grid-cols-2 gap-3">
            <button
              type="button"
              @click="verifyStatus = 'Disetujui'"
              class="py-2 text-sm font-semibold rounded-lg border text-center transition-all"
              :class="verifyStatus === 'Disetujui' ? 'bg-emerald-600 text-white border-emerald-600' : 'bg-slate-50 text-slate-700 border-slate-200'"
            >
              ✓ Setujui
            </button>
            <button
              type="button"
              @click="verifyStatus = 'Ditolak'"
              class="py-2 text-sm font-semibold rounded-lg border text-center transition-all"
              :class="verifyStatus === 'Ditolak' ? 'bg-rose-600 text-white border-rose-600' : 'bg-slate-50 text-slate-700 border-slate-200'"
            >
              ✕ Tolak (Revisi)
            </button>
          </div>
        </div>

        <div v-if="verifyStatus === 'Ditolak'">
          <label class="block text-xs font-semibold text-rose-700 mb-1">Catatan Revisi untuk Pegawai *</label>
          <textarea
            v-model="verifyCatatan"
            rows="3"
            placeholder="Contoh: Dokumen buram, harap scan ulang dengan resolusi lebih tinggi..."
            class="w-full text-xs p-2.5 border border-rose-300 rounded-lg outline-none focus:ring-2 focus:ring-rose-400"
          ></textarea>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button
            @click="verifyTarget = null"
            class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-lg"
          >
            Batal
          </button>
          <button
            @click="submitVerifikasi"
            :disabled="isSubmittingVerify || (verifyStatus === 'Ditolak' && !verifyCatatan.trim())"
            class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:opacity-50 text-white text-xs font-semibold rounded-lg"
          >
            Simpan Keputusan
          </button>
        </div>
      </div>
    </div>

    <!-- Preview Modal -->
    <DokumenViewerModal
      :is-open="previewModalOpen"
      :title="previewTitle"
      :file-name="previewFileName"
      :file-url="previewUrl"
      @close="previewModalOpen = false"
    />
  </div>
</template>
