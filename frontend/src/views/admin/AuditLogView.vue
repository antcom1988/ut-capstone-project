<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import api from '@/api/axios'
import { apiErrorMessage } from '@/utils/validation'
import type { AuditLog } from '@/types'

const logs = ref<AuditLog[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const limit = ref(20)
const total = ref(0)
const pages = ref(1)
const filters = ref({ from: '', to: '', user: '', action: '' })

const firstRow = computed(() => total.value ? (page.value - 1) * limit.value + 1 : 0)
const lastRow = computed(() => Math.min(page.value * limit.value, total.value))

async function load(targetPage = page.value) {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, string | number> = { page: targetPage, limit: limit.value }
    for (const [key, value] of Object.entries(filters.value)) if (value) params[key] = value
    const response = await api.get('/admin/audit-logs', { params })
    logs.value = response.data.data || []
    page.value = response.data.pagination?.page || targetPage
    total.value = response.data.pagination?.total || 0
    pages.value = Math.max(response.data.pagination?.pages || 1, 1)
  } catch (e) {
    error.value = apiErrorMessage(e, 'Gagal memuat audit log.')
  } finally {
    loading.value = false
  }
}

function applyFilters() { load(1) }
function resetFilters() {
  filters.value = { from: '', to: '', user: '', action: '' }
  load(1)
}
function changePage(next: number) {
  if (next >= 1 && next <= pages.value && next !== page.value) load(next)
}
function changeLimit() { load(1) }

onMounted(() => load(1))
</script>

<template>
  <section class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold">Audit Log</h1>
      <p class="text-sm text-slate-500">Riwayat aktivitas penting sistem.</p>
    </div>

    <form @submit.prevent="applyFilters" class="grid grid-cols-1 gap-3 rounded-xl border bg-white p-4 md:grid-cols-2 xl:grid-cols-5">
      <label class="text-xs font-semibold text-slate-600">Mulai
        <input v-model="filters.from" type="datetime-local" class="mt-1 w-full rounded-lg border p-2 text-sm font-normal" />
      </label>
      <label class="text-xs font-semibold text-slate-600">Selesai
        <input v-model="filters.to" type="datetime-local" class="mt-1 w-full rounded-lg border p-2 text-sm font-normal" />
      </label>
      <label class="text-xs font-semibold text-slate-600">Pengguna
        <input v-model.trim="filters.user" type="search" placeholder="Cari username" class="mt-1 w-full rounded-lg border p-2 text-sm font-normal" />
      </label>
      <label class="text-xs font-semibold text-slate-600">Aksi
        <input v-model.trim="filters.action" type="search" placeholder="Contoh: LOGIN" class="mt-1 w-full rounded-lg border p-2 text-sm font-normal" />
      </label>
      <div class="flex items-end gap-2">
        <button type="submit" :disabled="loading" class="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50">Filter</button>
        <button type="button" @click="resetFilters" :disabled="loading" class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-700 disabled:opacity-50">Reset</button>
      </div>
    </form>

    <div v-if="error" role="alert" class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-amber-800">{{ error }}</div>

    <div class="overflow-x-auto rounded-xl border bg-white">
      <table class="w-full text-sm">
        <thead class="bg-slate-50"><tr><th class="p-3 text-left">Waktu</th><th class="p-3 text-left">Pengguna</th><th class="p-3 text-left">Aksi</th><th class="p-3 text-left">Rincian</th><th class="p-3 text-left">IP</th></tr></thead>
        <tbody>
          <tr v-for="x in logs" :key="x.id" class="border-t"><td class="whitespace-nowrap p-3">{{ new Date(x.created_at).toLocaleString('id-ID') }}</td><td class="p-3">{{ x.username || '-' }}</td><td class="p-3 font-semibold">{{ x.action }}</td><td class="max-w-xl break-words p-3">{{ x.details || '-' }}</td><td class="p-3 font-mono">{{ x.ip_address || '-' }}</td></tr>
          <tr v-if="loading"><td colspan="5" class="p-8 text-center text-slate-400">Memuat audit log...</td></tr>
          <tr v-else-if="!logs.length"><td colspan="5" class="p-8 text-center text-slate-400">Tidak ada data audit pada filter ini.</td></tr>
        </tbody>
      </table>
    </div>

    <div class="flex flex-col gap-3 text-sm text-slate-600 sm:flex-row sm:items-center sm:justify-between">
      <div>Menampilkan {{ firstRow }}–{{ lastRow }} dari {{ total }} data</div>
      <div class="flex items-center gap-2">
        <select v-model.number="limit" @change="changeLimit" class="rounded-lg border bg-white px-2 py-1.5"><option :value="10">10 / halaman</option><option :value="20">20 / halaman</option><option :value="50">50 / halaman</option><option :value="100">100 / halaman</option></select>
        <button @click="changePage(page - 1)" :disabled="loading || page <= 1" class="rounded-lg border bg-white px-3 py-1.5 disabled:opacity-40">Sebelumnya</button>
        <span>Halaman {{ page }} / {{ pages }}</span>
        <button @click="changePage(page + 1)" :disabled="loading || page >= pages" class="rounded-lg border bg-white px-3 py-1.5 disabled:opacity-40">Berikutnya</button>
      </div>
    </div>
  </section>
</template>
