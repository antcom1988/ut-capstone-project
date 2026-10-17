<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import type { ArsipKinerja } from '@/types'
import DokumenViewerModal from '@/components/DokumenViewerModal.vue'
import { MAX_KINERJA_YEAR, validateKinerjaYear, validateFile } from '@/utils/validation'

const authStore = useAuthStore()
const kinerjaList = ref<ArsipKinerja[]>([])
const isLoading = ref(true)
const loadError = ref('')

// Modal Unggah Kinerja
const isModalOpen = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const form = ref({
  tahun_periode: new Date().getFullYear(),
  jenis_dokumen: 'SKP' as 'SKP' | 'PK',
})
const kinerjaFile = ref<File | null>(null)

// Preview Modal
const previewModalOpen = ref(false)
const previewTitle = ref('')
const previewUrl = ref('')
const previewFileName = ref('')

async function fetchKinerja() {
  isLoading.value = true
  loadError.value = ''
  try {
    const nip = authStore.user?.nip
    if (!nip) return
    const res = await api.get(`/kinerja/${nip}`)
    if (res.data.success) kinerjaList.value = res.data.data
  } catch (err) {
    loadError.value = 'Gagal memuat atau memperbarui data. Silakan coba lagi.'
  } finally {
    isLoading.value = false
  }
}

function openModal() {
  errorMessage.value = ''
  form.value = { tahun_periode: new Date().getFullYear(), jenis_dokumen: 'SKP' }
  kinerjaFile.value = null
  isModalOpen.value = true
}

function onFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) kinerjaFile.value = target.files[0]
}

async function submitKinerja() {
  if (!kinerjaFile.value) return
  errorMessage.value = validateKinerjaYear(form.value.tahun_periode) || validateFile(kinerjaFile.value, 1)
  if (errorMessage.value) return
  isSaving.value = true

  const formData = new FormData()
  formData.append('tahun_periode', form.value.tahun_periode.toString())
  formData.append('jenis_dokumen', form.value.jenis_dokumen)
  formData.append('file', kinerjaFile.value)

  try {
    const res = await api.post('/kinerja/upload', formData)
    if (res.data.success) {
      isModalOpen.value = false
      await fetchKinerja()
    }
  } catch (err: any) {
    errorMessage.value = err.response?.data?.message || 'Gagal mengunggah dokumen kinerja'
  } finally {
    isSaving.value = false
  }
}

function openPreview(k: ArsipKinerja) {
  previewTitle.value = `${k.jenis_dokumen} Tahun ${k.tahun_periode}`
  previewFileName.value = k.file_name || ''
  previewUrl.value = `/api/kinerja/preview/${k.id}`
  previewModalOpen.value = true
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(fetchKinerja)
</script>

<template>
  <div class="space-y-6">
    <p v-if="isLoading" role="status">Memuat data...</p>
    <p v-if="loadError" role="alert" class="p-3 bg-rose-50 text-rose-700 rounded-lg">{{ loadError }}</p>
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">Arsip Kinerja SKP & Perjanjian Kinerja</h1>
        <p class="text-sm text-slate-500">Kelola dokumen evaluasi kinerja tahunan Anda (SKP dan Perjanjian Kinerja / PK).</p>
      </div>
      <button
        @click="openModal"
        class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white rounded-xl text-sm font-semibold shadow-sm transition-colors"
      >
        ➕ Unggah Dokumen Kinerja
      </button>
    </div>

    <!-- Grid Kartu Kinerja Per Tahun -->
    <div v-if="!isLoading && kinerjaList.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div
        v-for="k in kinerjaList"
        :key="k.id"
        class="bg-white rounded-2xl border border-slate-200 shadow-sm p-5 flex flex-col justify-between hover:border-slate-300 transition-colors"
      >
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <span
              class="px-2.5 py-1 rounded-lg font-bold text-xs"
              :class="k.jenis_dokumen === 'SKP' ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'"
            >
              {{ k.jenis_dokumen === 'SKP' ? 'Sasaran Kinerja Pegawai' : 'Perjanjian Kinerja' }}
            </span>
            <span class="text-xl font-bold text-slate-700 font-mono">{{ k.tahun_periode }}</span>
          </div>
          <p class="text-[11px] text-slate-500 font-mono truncate" title="{{ k.file_name }}">
            📄 {{ k.file_name }}
          </p>
          <p class="text-[11px] text-slate-400">Diunggah: {{ formatDate(k.uploaded_at) }}</p>
        </div>
        <button
          @click="openPreview(k)"
          class="mt-4 w-full py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-semibold transition-colors"
        >
          👁️ Lihat Dokumen
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!isLoading && kinerjaList.length === 0" class="bg-white rounded-2xl border border-slate-200 shadow-sm py-16 text-center text-slate-400">
      <div class="text-4xl mb-3">📊</div>
      <p class="font-semibold text-slate-500">Belum ada arsip dokumen kinerja</p>
      <p class="text-sm mt-1">Klik tombol "Unggah Dokumen Kinerja" untuk menambahkan SKP atau PK tahunan Anda.</p>
    </div>

    <!-- Modal Unggah Kinerja -->
    <div v-if="isModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="bg-white rounded-2xl shadow-xl max-w-md w-full p-6 space-y-4">
        <div class="flex items-center justify-between border-b pb-3">
          <h3 class="text-base font-bold text-slate-800">Unggah Dokumen Kinerja</h3>
          <button @click="isModalOpen = false" class="text-slate-400 hover:text-slate-600 text-lg">✕</button>
        </div>

        <div v-if="errorMessage" class="p-3 bg-rose-50 border border-rose-200 text-rose-700 rounded-lg text-xs">
          {{ errorMessage }}
        </div>

        <form @submit.prevent="submitKinerja" class="space-y-3 text-xs">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Jenis Dokumen *</label>
              <select v-model="form.jenis_dokumen" class="w-full p-2.5 border border-slate-300 rounded-lg outline-none">
                <option value="SKP">SKP (Sasaran Kinerja Pegawai)</option>
                <option value="PK">PK (Perjanjian Kinerja)</option>
              </select>
            </div>
            <div>
              <label class="block font-semibold text-slate-600 mb-1">Tahun Periode *</label>
              <input
                v-model.number="form.tahun_periode" type="number" required min="2000" :max="MAX_KINERJA_YEAR"
                class="w-full p-2.5 border border-slate-300 rounded-lg outline-none font-mono"
              />
            </div>
          </div>

          <div>
            <label class="block font-semibold text-slate-600 mb-1">File Dokumen PDF / JPG / PNG (Maks 1 MB) *</label>
            <div class="border-2 border-dashed border-slate-200 rounded-xl p-5 text-center hover:border-primary-300 transition-colors">
              <p class="text-slate-400 mb-2 text-xs">Klik untuk memilih atau seret file ke sini</p>
              <input type="file" accept=".pdf,.jpg,.jpeg,.png" required @change="onFileChange" class="text-xs text-slate-500" />
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-2 border-t">
            <button type="button" @click="isModalOpen = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg font-semibold">Batal</button>
            <button type="submit" :disabled="isSaving || !kinerjaFile" class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-semibold disabled:opacity-50">
              {{ isSaving ? 'Mengunggah...' : 'Simpan Dokumen' }}
            </button>
          </div>
        </form>
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
