<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import type { RiwayatDiklat } from '@/types'
import DokumenViewerModal from '@/components/DokumenViewerModal.vue'
import { validateDateRange, validateFile } from '@/utils/validation'

const authStore = useAuthStore()
const diklatList = ref<RiwayatDiklat[]>([])
const isLoading = ref(true)
const loadError = ref('')

// Modal Tambah Diklat
const isModalOpen = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const form = ref({
  nama_diklat: '',
  penyelenggara: '',
  jumlah_jp: 30,
  tanggal_mulai: '',
  tanggal_selesai: '',
})
const sertifikatFile = ref<File | null>(null)

// Preview Modal
const previewModalOpen = ref(false)
const previewTitle = ref('')
const previewUrl = ref('')
const previewFileName = ref('')

async function fetchDiklat() {
  isLoading.value = true
  loadError.value = ''
  try {
    const nip = authStore.user?.nip
    if (!nip) return
    const res = await api.get(`/diklat/${nip}`)
    if (res.data.success) diklatList.value = res.data.data
  } catch (err) {
    loadError.value = 'Gagal memuat atau memperbarui data. Silakan coba lagi.'
  } finally {
    isLoading.value = false
  }
}

const editingDiklat = ref<RiwayatDiklat | null>(null)
const deletingDiklat = ref<number | null>(null)
function openEditDiklat(d: RiwayatDiklat) {
 openModal()
 editingDiklat.value = d
 form.value = { nama_diklat: d.nama_diklat, penyelenggara: d.penyelenggara, jumlah_jp: d.jumlah_jp, tanggal_mulai: d.tanggal_mulai.substring(0,10), tanggal_selesai: d.tanggal_selesai.substring(0,10) }
}
async function deleteDiklat(d: RiwayatDiklat) {
 if (!confirm(`Hapus riwayat diklat "${d.nama_diklat}" beserta sertifikatnya?`)) return
 deletingDiklat.value = d.id; loadError.value = ''
 try { await api.delete(`/diklat/${d.id}`); await fetchDiklat() }
 catch (e: any) { loadError.value = e.response?.data?.message || 'Gagal menghapus diklat.' }
 finally { deletingDiklat.value = null }
}
function openModal() {
  editingDiklat.value = null
  errorMessage.value = ''
  form.value = { nama_diklat: '', penyelenggara: '', jumlah_jp: 30, tanggal_mulai: '', tanggal_selesai: '' }
  sertifikatFile.value = null
  isModalOpen.value = true
}

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) sertifikatFile.value = target.files[0]
}

async function submitDiklat() {
  errorMessage.value = validateDateRange(form.value.tanggal_mulai, form.value.tanggal_selesai) || (sertifikatFile.value ? validateFile(sertifikatFile.value, 2) : editingDiklat.value?.file_name ? '' : 'Sertifikat wajib diunggah (maks 2 MB).')
  if (errorMessage.value) return
  isSaving.value = true

  const formData = new FormData()
  formData.append('nama_diklat', form.value.nama_diklat)
  formData.append('penyelenggara', form.value.penyelenggara)
  formData.append('jumlah_jp', form.value.jumlah_jp.toString())
  formData.append('tanggal_mulai', form.value.tanggal_mulai)
  formData.append('tanggal_selesai', form.value.tanggal_selesai)
  if (sertifikatFile.value) {
    formData.append('file_sertifikat', sertifikatFile.value)
  }

  try {
    const res = editingDiklat.value ? await api.put(`/diklat/${editingDiklat.value.id}`, formData) : await api.post('/diklat', formData)
    if (res.data.success) {
      isModalOpen.value = false
      await fetchDiklat()
    }
  } catch (err: any) {
    errorMessage.value = err.response?.data?.message || 'Gagal menyimpan data diklat'
  } finally {
    isSaving.value = false
  }
}

function openPreview(d: RiwayatDiklat) {
  previewTitle.value = `Sertifikat: ${d.nama_diklat}`
  previewFileName.value = d.file_name || ''
  previewUrl.value = `/api/diklat/preview/${d.id}`
  previewModalOpen.value = true
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(fetchDiklat)
</script>

<template>
  <div class="space-y-6">
    <p v-if="isLoading" role="status">Memuat data...</p>
    <p v-if="loadError" role="alert" class="p-3 bg-rose-50 text-rose-700 rounded-lg">{{ loadError }}</p>
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">Rekam Jejak Diklat & Kompetensi</h1>
        <p class="text-sm text-slate-500">Catat riwayat pelatihan teknis/fungsional beserta sertifikat bukti keikutsertaan.</p>
      </div>
      <button
        @click="openModal"
        class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white rounded-xl text-sm font-semibold shadow-sm transition-colors"
      >
        ➕ Tambah Diklat
      </button>
    </div>

    <!-- Tabel Diklat -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-slate-600 text-xs font-semibold uppercase border-b border-slate-200">
            <tr>
              <th class="px-6 py-3.5">Nama Diklat / Pelatihan</th>
              <th class="px-6 py-3.5">Penyelenggara</th>
              <th class="px-6 py-3.5 text-center">JP</th>
              <th class="px-6 py-3.5">Periode</th>
              <th class="px-6 py-3.5 text-right">Sertifikat</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700">
            <tr v-for="d in diklatList" :key="d.id" class="hover:bg-slate-50/80 transition-colors">
              <td class="px-6 py-4 font-semibold text-slate-900">{{ d.nama_diklat }}</td>
              <td class="px-6 py-4 text-sm text-slate-600">{{ d.penyelenggara }}</td>
              <td class="px-6 py-4 text-center font-semibold text-slate-700">{{ d.jumlah_jp }}<span class="font-normal text-slate-400 text-xs"> JP</span></td>
              <td class="px-6 py-4 text-xs font-mono text-slate-600">
                {{ formatDate(d.tanggal_mulai) }} – {{ formatDate(d.tanggal_selesai) }}
              </td>
              <td class="px-6 py-4 text-right">
                <button @click="openEditDiklat(d)" class="mr-2 text-primary-700">Edit</button><button @click="deleteDiklat(d)" :disabled="deletingDiklat !== null" class="mr-2 text-rose-700">Hapus</button>
                <button
                  v-if="d.file_name"
                  @click="openPreview(d)"
                  class="px-3 py-1.5 bg-primary-50 hover:bg-primary-100 text-primary-700 border border-primary-200 rounded-lg text-xs font-semibold transition-colors inline-flex items-center gap-1"
                >
                  👁️ Lihat Sertifikat
                </button>
                <span v-else class="text-xs text-slate-400 italic">Tidak ada berkas</span>
              </td>
            </tr>
            <tr v-if="!isLoading && diklatList.length === 0">
              <td colspan="5" class="px-6 py-10 text-center text-slate-400">
                <div class="text-3xl mb-2">🎓</div>
                Belum ada rekaman diklat yang ditambahkan.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal Tambah Diklat -->
    <div v-if="isModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="bg-white rounded-2xl shadow-xl max-w-lg w-full p-6 space-y-4">
        <div class="flex items-center justify-between border-b pb-3">
          <h3 class="text-base font-bold text-slate-800">{{ editingDiklat ? 'Edit Rekaman Diklat' : 'Tambah Rekaman Diklat' }}</h3>
          <button @click="isModalOpen = false" class="text-slate-400 hover:text-slate-600 text-lg">✕</button>
        </div>

        <div v-if="errorMessage" class="p-3 bg-rose-50 border border-rose-200 text-rose-700 rounded-lg text-xs">
          {{ errorMessage }}
        </div>

        <form @submit.prevent="submitDiklat" class="space-y-3 text-xs">
          <div>
            <label class="block font-semibold text-slate-600 mb-1">Nama Diklat / Pelatihan *</label>
            <input
              v-model="form.nama_diklat" type="text" required
              placeholder="Contoh: Diklat Pengelolaan Keuangan Daerah"
              class="w-full p-2.5 border border-slate-300 rounded-lg outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">Penyelenggara *</label>
            <input
              v-model="form.penyelenggara" type="text" required
              placeholder="Contoh: BPSDM Provinsi Jawa Timur"
              class="w-full p-2.5 border border-slate-300 rounded-lg outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">Jumlah Jam Pelajaran (JP) *</label>
            <input
              v-model.number="form.jumlah_jp" type="number" required min="1"
              class="w-full p-2.5 border border-slate-300 rounded-lg outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Tanggal Mulai *</label>
              <input v-model="form.tanggal_mulai" type="date" required class="w-full p-2.5 border border-slate-300 rounded-lg" />
            </div>
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Tanggal Selesai *</label>
              <input v-model="form.tanggal_selesai" type="date" required class="w-full p-2.5 border border-slate-300 rounded-lg" />
            </div>
          </div>
          <div>
            <label class="block font-semibold text-slate-600 mb-1">File Sertifikat (PDF / JPG / PNG, maks 2 MB)</label>
            <p v-if="editingDiklat?.file_name" class="text-slate-500 mb-2">Sertifikat saat ini: {{ editingDiklat.file_name }}. Kosongkan untuk mempertahankan.</p>
            <div class="border-2 border-dashed border-slate-200 rounded-xl p-4 text-center hover:border-primary-300 transition-colors">
              <input type="file" accept=".pdf,.jpg,.jpeg,.png" :required="!editingDiklat?.file_name" @change="onFileChange" class="text-xs text-slate-500" />
            </div>
          </div>
          <div class="flex justify-end gap-2 pt-2 border-t">
            <button type="button" @click="isModalOpen = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg font-semibold">Batal</button>
            <button type="submit" :disabled="isSaving" class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-semibold disabled:opacity-50">
              {{ isSaving ? 'Menyimpan...' : 'Simpan Diklat' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Preview Sertifikat -->
    <DokumenViewerModal
      :is-open="previewModalOpen"
      :title="previewTitle"
      :file-name="previewFileName"
      :file-url="previewUrl"
      @close="previewModalOpen = false"
    />
  </div>
</template>
