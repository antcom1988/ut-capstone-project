<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api/axios'
import type { DashboardSummary } from '@/types'

const summary = ref<DashboardSummary | null>(null)
const thresholdMonths = ref(3)
const isUpdatingThreshold = ref(false)
const isLoading = ref(true)
const loadError = ref('')
const feedback = ref('')

async function fetchDashboard() {
  isLoading.value = true
  loadError.value = ''
  try {
    const res = await api.get('/admin/dashboard')
    if (res.data.success) {
      summary.value = res.data.data
      thresholdMonths.value = res.data.data.threshold_months ?? thresholdMonths.value
    }
  } catch (err) {
    loadError.value = 'Gagal memuat atau memperbarui data. Silakan coba lagi.'
  } finally {
    isLoading.value = false
  }
}

async function updateThreshold() {
  isUpdatingThreshold.value = true
  try {
    await api.put('/admin/settings/alert-threshold', {
      threshold_months: thresholdMonths.value,
    })
    await fetchDashboard()
    feedback.value = 'Ambang peringatan berhasil diperbarui.'
  } catch (err: any) {
    feedback.value = err.response?.data?.message || 'Gagal memperbarui ambang peringatan.'
  } finally {
    isUpdatingThreshold.value = false
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(() => {
  fetchDashboard()
})
</script>

<template>
  <div class="space-y-6">
    <p v-if="isLoading" role="status">Memuat data...</p>
    <p v-if="loadError" role="alert" class="p-3 bg-rose-50 text-rose-700 rounded-lg">{{ loadError }}</p>
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">Dashboard Monitoring Kepegawaian</h1>
        <p class="text-sm text-slate-500">Sub Bagian Umum dan Kepegawaian – BPKAD Kabupaten Jember</p>
      </div>

      <!-- Quick Threshold Setting -->
      <div class="flex items-center gap-2 bg-white px-3.5 py-2 rounded-xl border border-slate-200 shadow-sm text-sm">
        <span class="text-slate-600 font-medium text-xs">Peringatan:</span>
        <select
          v-model="thresholdMonths"
          @change="updateThreshold"
          :disabled="isUpdatingThreshold"
          class="bg-slate-50 border border-slate-300 rounded-md px-2 py-1 text-xs font-semibold text-slate-700 outline-none"
        >
          <option :value="1">1 Bulan Sebelum</option>
          <option :value="2">2 Bulan Sebelum</option>
          <option :value="3">3 Bulan Sebelum (Standar)</option>
          <option :value="6">6 Bulan Sebelum</option>
        </select>
      </div>
    </div>

    <p v-if="feedback" role="status" class="p-3 rounded-lg bg-slate-50 border text-sm text-slate-700">{{ feedback }}</p>

    <!-- Metrik Utama (KPI Cards) -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total Pegawai -->
      <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center font-bold text-xl">
          👥
        </div>
        <div>
          <span class="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Pegawai</span>
          <div class="text-2xl font-bold text-slate-800">{{ summary?.total_pegawai ?? 0 }}</div>
        </div>
      </div>

      <!-- Berkas Lengkap % -->
      <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold text-xl">
          📋
        </div>
        <div>
          <span class="text-xs font-medium text-slate-400 uppercase tracking-wider">Kelengkapan Berkas</span>
          <div class="text-2xl font-bold text-slate-800">
            {{ summary?.berkas_lengkap_persen.toFixed(1) ?? 0 }}%
          </div>
        </div>
      </div>

      <!-- Menunggu Verifikasi -->
      <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold text-xl">
          ⏳
        </div>
        <div>
          <span class="text-xs font-medium text-slate-400 uppercase tracking-wider">Menunggu Verifikasi</span>
          <div class="text-2xl font-bold text-amber-600">
            {{ summary?.menunggu_verifikasi ?? 0 }}
          </div>
        </div>
      </div>

      <!-- Alert Mendekati -->
      <div class="bg-white p-5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center font-bold text-xl">
          🔔
        </div>
        <div>
          <span class="text-xs font-medium text-slate-400 uppercase tracking-wider">KGB/KP Jatuh Tempo</span>
          <div class="text-2xl font-bold text-rose-600">
            {{ summary?.alert_mendekati ?? 0 }}
          </div>
        </div>
      </div>
    </div>

    <!-- Tabel Control Alert System KGB & KP -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <div>
          <h2 class="text-base font-semibold text-slate-800">Daftar Peringatan KGB & Kenaikan Pangkat</h2>
          <p class="text-xs text-slate-500">Kalkulasi otomatis berdasarkan TMT KGB dan TMT Pangkat Terakhir</p>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-slate-600 text-xs font-semibold uppercase border-b border-slate-200">
            <tr>
              <th class="px-6 py-3.5">Pegawai</th>
              <th class="px-6 py-3.5">Bidang / Pangkat</th>
              <th class="px-6 py-3.5">Jenis</th>
              <th class="px-6 py-3.5">TMT Acuan</th>
              <th class="px-6 py-3.5">Jatuh Tempo</th>
              <th class="px-6 py-3.5">Status Peringatan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700">
            <tr v-for="item in summary?.alerts" :key="item.nip + item.tipe_alert" class="hover:bg-slate-50/80 transition-colors">
              <td class="px-6 py-3.5">
                <div class="font-medium text-slate-900">{{ item.nama_lengkap }}</div>
                <div class="text-xs text-slate-400 font-mono">{{ item.nip }}</div>
              </td>
              <td class="px-6 py-3.5">
                <div>{{ item.bidang }}</div>
                <div class="text-xs text-slate-500">{{ item.pangkat_golongan }}</div>
              </td>
              <td class="px-6 py-3.5">
                <span class="px-2.5 py-0.5 rounded font-bold text-xs" :class="item.tipe_alert === 'KGB' ? 'bg-indigo-50 text-indigo-700' : 'bg-purple-50 text-purple-700'">
                  {{ item.tipe_alert }}
                </span>
              </td>
              <td class="px-6 py-3.5 text-xs font-mono">
                {{ formatDate(item.tmt_dasar) }}
              </td>
              <td class="px-6 py-3.5 text-xs font-mono font-medium">
                {{ formatDate(item.tanggal_jatuh_tempo) }}
              </td>
              <td class="px-6 py-3.5">
                <span
                  v-if="item.status === 'SEGERA'"
                  class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-rose-100 text-rose-800"
                >
                  <span class="w-2 h-2 rounded-full bg-rose-600 animate-pulse"></span>
                  {{ item.sisa_hari < 0 ? `Terlambat ${Math.abs(item.sisa_hari)} hari` : item.sisa_hari === 0 ? 'Jatuh tempo hari ini' : `Segera (${item.sisa_hari} hari)` }}
                </span>
                <span
                  v-else-if="item.status === 'DIINGATKAN'"
                  class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-amber-100 text-amber-800"
                >
                  <span class="w-2 h-2 rounded-full bg-amber-600"></span>
                  Diingatkan ({{ item.sisa_bulan.toFixed(1) }} bln)
                </span>
                <span
                  v-else
                  class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800"
                >
                  <span class="w-2 h-2 rounded-full bg-emerald-600"></span>
                  Aman
                </span>
              </td>
            </tr>
            <tr v-if="!summary?.alerts || summary.alerts.length === 0">
              <td colspan="6" class="px-6 py-8 text-center text-slate-400">
                Tidak ada data pegawai.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
