<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api/axios'
import type { Pegawai } from '@/types'
import { validateContact, validateNip } from '@/utils/validation'

const pegawais = ref<Pegawai[]>([])
const searchQuery = ref('')
const selectedBidang = ref('')
const isLoading = ref(true)
const loadError = ref('')

// Modal State
const isModalOpen = ref(false)
const isEditMode = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')

const form = ref({
  nip: '',
  nama_lengkap: '',
  tempat_lahir: '',
  tanggal_lahir: '',
  jenis_kelamin: 'L',
  email: '',
  no_hp: '',
  status_pegawai: 'PNS',
  bidang: 'Bidang Akuntansi',
  jenis_jabatan: 'Fungsional',
  nama_jabatan: '',
  pangkat: 'Penata Muda',
  golongan: 'III/a',
  tmt_pengangkatan: '',
  tmt_pangkat_terakhir: '',
  tmt_kgb_terakhir: '',
  jenjang_pendidikan: 'S1',
  jurusan: '',
  institusi_pendidikan: '',
  password: '',
})

async function fetchPegawai() {
  isLoading.value = true
  loadError.value = ''
  try {
    const res = await api.get('/admin/pegawai')
    if (res.data.success) {
      pegawais.value = res.data.data
    }
  } catch (err) {
    loadError.value = 'Gagal memuat atau memperbarui data. Silakan coba lagi.'
  } finally {
    isLoading.value = false
  }
}

const filteredPegawai = computed(() => {
  return pegawais.value.filter((p) => {
    const matchesSearch =
      p.nama_lengkap.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      p.nip.includes(searchQuery.value) ||
      p.nama_jabatan.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesBidang = selectedBidang.value ? p.bidang === selectedBidang.value : true
    return matchesSearch && matchesBidang
  })
})

function openAddModal() {
  isEditMode.value = false
  errorMessage.value = ''
  form.value = {
    nip: '',
    nama_lengkap: '',
    tempat_lahir: '',
    tanggal_lahir: '',
    jenis_kelamin: 'L',
    email: '',
    no_hp: '',
    status_pegawai: 'PNS',
    bidang: 'Bidang Akuntansi',
    jenis_jabatan: 'Fungsional',
    nama_jabatan: '',
    pangkat: 'Penata Muda',
    golongan: 'III/a',
    tmt_pengangkatan: '',
    tmt_pangkat_terakhir: '',
    tmt_kgb_terakhir: '',
    jenjang_pendidikan: 'S1',
    jurusan: '',
    institusi_pendidikan: '',
    password: '',
  }
  isModalOpen.value = true
}

function openEditModal(p: Pegawai) {
  isEditMode.value = true
  errorMessage.value = ''
  form.value = {
    nip: p.nip,
    nama_lengkap: p.nama_lengkap,
    tempat_lahir: p.tempat_lahir || '',
    tanggal_lahir: p.tanggal_lahir ? p.tanggal_lahir.substring(0, 10) : '',
    jenis_kelamin: p.jenis_kelamin || 'L',
    email: p.email || '',
    no_hp: p.no_hp || '',
    status_pegawai: p.status_pegawai,
    bidang: p.bidang,
    jenis_jabatan: p.jenis_jabatan,
    nama_jabatan: p.nama_jabatan,
    pangkat: p.pangkat,
    golongan: p.golongan,
    tmt_pengangkatan: p.tmt_pengangkatan ? p.tmt_pengangkatan.substring(0, 10) : '',
    tmt_pangkat_terakhir: p.tmt_pangkat_terakhir ? p.tmt_pangkat_terakhir.substring(0, 10) : '',
    tmt_kgb_terakhir: p.tmt_kgb_terakhir ? p.tmt_kgb_terakhir.substring(0, 10) : '',
    jenjang_pendidikan: p.jenjang_pendidikan || 'S1',
    jurusan: p.jurusan || '',
    institusi_pendidikan: p.institusi_pendidikan || '',
    password: '',
  }
  isModalOpen.value = true
}

async function savePegawai() {
  errorMessage.value = validateNip(form.value.nip) || validateContact(form.value.email, form.value.no_hp)
  if (!isEditMode.value && form.value.password.length < 8) errorMessage.value = 'Kata sandi awal wajib minimal 8 karakter.'
  if (errorMessage.value) return
  isSaving.value = true
  try {
    if (isEditMode.value) {
      const res = await api.put(`/admin/pegawai/${form.value.nip}`, form.value)
      if (res.data.success) {
        isModalOpen.value = false
        await fetchPegawai()
      }
    } else {
      const res = await api.post('/admin/pegawai', { ...form.value, must_change_password: true })
      if (res.data.success) {
        isModalOpen.value = false
        await fetchPegawai()
      }
    }
  } catch (err: any) {
    errorMessage.value = err.response?.data?.message || 'Gagal menyimpan data pegawai'
  } finally {
    isSaving.value = false
  }
}

async function toggleStatus(p: Pegawai) {
  const confirmMsg = p.is_active
    ? `Apakah Anda yakin ingin menonaktifkan pegawai ${p.nama_lengkap}?`
    : `Aktifkan kembali pegawai ${p.nama_lengkap}?`
  if (!confirm(confirmMsg)) return

  try {
    const res = await api.put(`/admin/pegawai/${p.nip}/toggle-status`)
    if (res.data.success) {
      p.is_active = res.data.is_active
    }
  } catch (err) {
    loadError.value = 'Gagal memuat atau memperbarui data. Silakan coba lagi.'
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(() => {
  fetchPegawai()
})
</script>

<template>
  <div class="space-y-6">
    <p v-if="isLoading" role="status">Memuat data...</p>
    <p v-if="loadError" role="alert" class="p-3 bg-rose-50 text-rose-700 rounded-lg">{{ loadError }}</p>
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">Master Data Pegawai</h1>
        <p class="text-sm text-slate-500">Kelola informasi personalia, jabatan, pendidikan, dan acuan TMT KGB/KP</p>
      </div>
      <button
        @click="openAddModal"
        class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white rounded-xl text-sm font-semibold shadow-sm transition-colors"
      >
        <span>➕ Tambah Pegawai Baru</span>
      </button>
    </div>

    <!-- Filter & Pencarian -->
    <div class="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex flex-col md:flex-row gap-3 items-center justify-between">
      <div class="w-full md:w-80">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari Nama / NIP / Jabatan..."
          class="w-full px-3.5 py-2 text-sm border border-slate-300 rounded-lg outline-none focus:ring-2 focus:ring-primary-500"
        />
      </div>
      <div class="w-full md:w-auto flex items-center gap-2">
        <label class="text-xs text-slate-500 font-medium">Bidang:</label>
        <select
          v-model="selectedBidang"
          class="px-3 py-2 text-xs border border-slate-300 rounded-lg outline-none bg-slate-50 font-medium text-slate-700"
        >
          <option value="">Semua Bidang</option>
          <option value="Sekretariat">Sekretariat</option>
          <option value="Bidang Anggaran">Bidang Anggaran</option>
          <option value="Bidang Perbendaharaan">Bidang Perbendaharaan</option>
          <option value="Bidang Akuntansi">Bidang Akuntansi</option>
          <option value="Bidang Aset Daerah">Bidang Aset Daerah</option>
        </select>
      </div>
    </div>

    <!-- Tabel Data Pegawai -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-slate-600 text-xs font-semibold uppercase border-b border-slate-200">
            <tr>
              <th class="px-6 py-3.5">Pegawai</th>
              <th class="px-6 py-3.5">Status & Jabatan</th>
              <th class="px-6 py-3.5">Pangkat / Gol.</th>
              <th class="px-6 py-3.5">TMT Pangkat</th>
              <th class="px-6 py-3.5">TMT KGB</th>
              <th class="px-6 py-3.5">Status</th>
              <th class="px-6 py-3.5 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700">
            <tr
              v-for="p in filteredPegawai"
              :key="p.nip"
              class="hover:bg-slate-50/80 transition-colors"
              :class="{ 'opacity-60 bg-slate-50/50': !p.is_active }"
            >
              <td class="px-6 py-3.5">
                <div class="font-bold text-slate-900">{{ p.nama_lengkap }}</div>
                <div class="text-xs text-slate-400 font-mono">NIP: {{ p.nip }}</div>
                <div class="text-[11px] text-slate-500 mt-0.5">{{ p.email || '-' }} • {{ p.no_hp || '-' }}</div>
              </td>
              <td class="px-6 py-3.5">
                <div class="text-xs font-semibold px-2 py-0.5 rounded inline-block mb-1" :class="p.status_pegawai === 'PNS' ? 'bg-blue-50 text-blue-700' : 'bg-emerald-50 text-emerald-700'">
                  {{ p.status_pegawai }} • {{ p.jenis_jabatan }}
                </div>
                <div class="font-medium text-slate-800 text-xs">{{ p.nama_jabatan }}</div>
                <div class="text-[11px] text-slate-400">{{ p.bidang }}</div>
              </td>
              <td class="px-6 py-3.5">
                <div class="font-semibold text-slate-800 text-xs">{{ p.golongan }}</div>
                <div class="text-xs text-slate-500">{{ p.pangkat }}</div>
              </td>
              <td class="px-6 py-3.5 text-xs font-mono">
                {{ formatDate(p.tmt_pangkat_terakhir) }}
              </td>
              <td class="px-6 py-3.5 text-xs font-mono">
                {{ formatDate(p.tmt_kgb_terakhir) }}
              </td>
              <td class="px-6 py-3.5">
                <span
                  class="px-2.5 py-1 rounded-full text-xs font-semibold"
                  :class="p.is_active ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-600'"
                >
                  {{ p.is_active ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="px-6 py-3.5 text-right space-x-2">
                <button
                  @click="openEditModal(p)"
                  class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-semibold transition-colors"
                >
                  Edit
                </button>
                <button
                  @click="toggleStatus(p)"
                  class="px-3 py-1 rounded-lg text-xs font-semibold transition-colors"
                  :class="p.is_active ? 'bg-rose-50 hover:bg-rose-100 text-rose-700' : 'bg-emerald-50 hover:bg-emerald-100 text-emerald-700'"
                >
                  {{ p.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
                </button>
              </td>
            </tr>
            <tr v-if="filteredPegawai.length === 0">
              <td colspan="7" class="px-6 py-8 text-center text-slate-400">
                Tidak ada data pegawai yang sesuai dengan filter.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal Form Tambah / Edit Pegawai -->
    <div v-if="isModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 overflow-y-auto">
      <div class="bg-white rounded-2xl shadow-2xl max-w-3xl w-full p-6 my-8 space-y-4">
        <div class="border-b border-slate-100 pb-3 flex items-center justify-between">
          <h3 class="text-lg font-bold text-slate-800">
            {{ isEditMode ? 'Edit Data Pegawai' : 'Tambah Pegawai Baru' }}
          </h3>
          <button @click="isModalOpen = false" class="text-slate-400 hover:text-slate-600">✕</button>
        </div>

        <div v-if="errorMessage" class="p-3 bg-rose-50 text-rose-700 border border-rose-200 rounded-lg text-xs">
          {{ errorMessage }}
        </div>

        <form @submit.prevent="savePegawai" class="space-y-4 text-xs">
          <!-- 1. IDENTITAS DIRI -->
          <div>
            <h4 class="font-bold text-slate-700 uppercase tracking-wider mb-2 text-[11px] bg-slate-100 p-2 rounded">
              1. Identitas Pribadi
            </h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div>
                <label class="block font-semibold text-slate-600 mb-1">NIP (18 Digit) *</label>
                <input
                  v-model="form.nip"
                  type="text"
                  required
                  :disabled="isEditMode"
                  maxlength="18"
                  placeholder="198507152010011012"
                  class="w-full p-2 border rounded-lg disabled:bg-slate-100 font-mono"
                />
              </div>
              <div class="md:col-span-2">
                <label class="block font-semibold text-slate-600 mb-1">Nama Lengkap (dengan Gelar) *</label>
                <input
                  v-model="form.nama_lengkap"
                  type="text"
                  required
                  placeholder="Budi Santoso, S.E., M.Ak."
                  class="w-full p-2 border rounded-lg"
                />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Tempat Lahir</label>
                <input v-model="form.tempat_lahir" type="text" class="w-full p-2 border rounded-lg" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Tanggal Lahir</label>
                <input v-model="form.tanggal_lahir" type="date" class="w-full p-2 border rounded-lg" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Jenis Kelamin</label>
                <select v-model="form.jenis_kelamin" class="w-full p-2 border rounded-lg">
                  <option value="L">Laki-laki</option>
                  <option value="P">Perempuan</option>
                </select>
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Email</label>
                <input v-model="form.email" type="email" placeholder="nama@jemberkab.go.id" class="w-full p-2 border rounded-lg" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">No. HP / WhatsApp</label>
                <input v-model="form.no_hp" type="text" placeholder="081234567890" class="w-full p-2 border rounded-lg" />
              </div>
              <div v-if="!isEditMode">
                <label class="block font-semibold text-slate-600 mb-1">Kata Sandi Akun</label>
                <input v-model="form.password" type="password" required minlength="8" autocomplete="new-password" placeholder="Minimal 8 karakter, wajib diganti saat masuk" class="w-full p-2 border rounded-lg" />
              </div>
            </div>
          </div>

          <!-- 2. DATA KEPEGAWAIAN & ACUAN TMT -->
          <div>
            <h4 class="font-bold text-slate-700 uppercase tracking-wider mb-2 text-[11px] bg-slate-100 p-2 rounded">
              2. Data Kepegawaian & Acuan Alert KGB / KP
            </h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Status Pegawai *</label>
                <select v-model="form.status_pegawai" class="w-full p-2 border rounded-lg">
                  <option value="PNS">PNS</option>
                  <option value="PPPK">PPPK</option>
                </select>
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Bidang Unit Kerja *</label>
                <select v-model="form.bidang" class="w-full p-2 border rounded-lg">
                  <option value="Sekretariat">Sekretariat</option>
                  <option value="Bidang Anggaran">Bidang Anggaran</option>
                  <option value="Bidang Perbendaharaan">Bidang Perbendaharaan</option>
                  <option value="Bidang Akuntansi">Bidang Akuntansi</option>
                  <option value="Bidang Aset Daerah">Bidang Aset Daerah</option>
                </select>
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Jenis Jabatan *</label>
                <select v-model="form.jenis_jabatan" class="w-full p-2 border rounded-lg">
                  <option value="Struktural">Struktural</option>
                  <option value="Fungsional">Fungsional</option>
                  <option value="Pelaksana">Pelaksana</option>
                </select>
              </div>
              <div class="md:col-span-2">
                <label class="block font-semibold text-slate-600 mb-1">Nama Jabatan *</label>
                <input v-model="form.nama_jabatan" type="text" required placeholder="Analis Keuangan Pusat dan Daerah Ahli Muda" class="w-full p-2 border rounded-lg" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Pangkat & Golongan *</label>
                <div class="flex gap-2">
                  <input v-model="form.pangkat" type="text" placeholder="Penata Tingkat I" class="w-2/3 p-2 border rounded-lg" />
                  <input v-model="form.golongan" type="text" placeholder="III/d" class="w-1/3 p-2 border rounded-lg font-bold" />
                </div>
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">TMT Pengangkatan Awal *</label>
                <input v-model="form.tmt_pengangkatan" type="date" required class="w-full p-2 border rounded-lg font-mono" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">TMT Pangkat Terakhir * (KP: +4 Thn)</label>
                <input v-model="form.tmt_pangkat_terakhir" type="date" required class="w-full p-2 border rounded-lg font-mono bg-purple-50/50 border-purple-200" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">TMT KGB Terakhir * (KGB: +2 Thn)</label>
                <input v-model="form.tmt_kgb_terakhir" type="date" required class="w-full p-2 border rounded-lg font-mono bg-indigo-50/50 border-indigo-200" />
              </div>
            </div>
          </div>

          <!-- 3. PENDIDIKAN -->
          <div>
            <h4 class="font-bold text-slate-700 uppercase tracking-wider mb-2 text-[11px] bg-slate-100 p-2 rounded">
              3. Pendidikan Terakhir
            </h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Jenjang</label>
                <select v-model="form.jenjang_pendidikan" class="w-full p-2 border rounded-lg">
                  <option value="SMA">SMA/SMK</option>
                  <option value="D3">D3</option>
                  <option value="S1">S1</option>
                  <option value="S2">S2</option>
                  <option value="S3">S3</option>
                </select>
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Jurusan</label>
                <input v-model="form.jurusan" type="text" placeholder="Magister Akuntansi" class="w-full p-2 border rounded-lg" />
              </div>
              <div>
                <label class="block font-semibold text-slate-600 mb-1">Institusi / Universitas</label>
                <input v-model="form.institusi_pendidikan" type="text" placeholder="Universitas Jember" class="w-full p-2 border rounded-lg" />
              </div>
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-3 border-t">
            <button
              type="button"
              @click="isModalOpen = false"
              class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg font-semibold"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSaving"
              class="px-5 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg font-semibold disabled:opacity-50"
            >
              {{ isSaving ? 'Menyimpan...' : 'Simpan Data' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
