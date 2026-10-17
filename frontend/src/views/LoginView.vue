<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/axios'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const isLoading = ref(false)
onMounted(() => { const notice = sessionStorage.getItem('sikapj_notice'); if (notice) { errorMessage.value = notice; sessionStorage.removeItem('sikapj_notice') } })

async function handleLogin() {
  errorMessage.value = ''
  isLoading.value = true
  try {
    const res = await api.post('/auth/login', {
      username: username.value,
      password: password.value,
    })

    if (res.data.success) {
      authStore.setAuth(res.data.data.token, res.data.data.user)
      if (res.data.data.user.must_change_password) {
        router.push('/change-password')
      } else if (res.data.data.user.role === 'ADMIN') {
        router.push(typeof router.currentRoute.value.query.redirect === 'string' ? router.currentRoute.value.query.redirect : '/admin/dashboard')
      } else {
        router.push(typeof router.currentRoute.value.query.redirect === 'string' ? router.currentRoute.value.query.redirect : '/pegawai/dashboard')
      }
    }
  } catch (err: any) {
    if (err.response?.data?.message) {
      errorMessage.value = err.response.data.message
    } else {
      errorMessage.value = 'Gagal menghubungi server backend. Pastikan server Go aktif.'
    }
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-100 to-slate-200 p-4">
    <div class="w-full max-w-md bg-white rounded-2xl shadow-xl border border-slate-200/80 p-8">
      <!-- Logo & Title -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary-50 text-primary-600 mb-3 border border-primary-100">
          <svg class="w-9 h-9" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
        <h1 class="text-2xl font-bold tracking-tight text-slate-800">SIKAP - J</h1>
        <p class="text-xs text-slate-500 mt-1 uppercase tracking-wider font-semibold">
          BPKAD Kabupaten Jember
        </p>
        <p class="text-xs text-slate-400 mt-0.5">
          Sistem Informasi Kepegawaian, Arsip, dan Pengingat
        </p>
      </div>

      <!-- Alert Error -->
      <div v-if="errorMessage" class="mb-5 p-3.5 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 text-sm flex items-start gap-2.5">
        <svg class="w-5 h-5 flex-shrink-0 text-rose-500 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Form Login -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-600 uppercase mb-1.5">NIP / Username</label>
          <input
            v-model="username"
            type="text"
            required
            placeholder="Masukkan 18 digit NIP atau username"
            class="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none text-sm transition-all text-slate-800"
          />
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-600 uppercase mb-1.5">Kata Sandi</label>
          <input
            v-model="password"
            type="password"
            required
            placeholder="Masukkan kata sandi"
            class="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none text-sm transition-all text-slate-800"
          />
        </div>

        <button
          type="submit"
          :disabled="isLoading"
          class="w-full mt-2 py-2.5 px-4 rounded-lg bg-primary-600 hover:bg-primary-700 text-white font-medium text-sm transition-colors shadow-sm disabled:opacity-50 flex items-center justify-center gap-2"
        >
          <svg v-if="isLoading" class="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span>{{ isLoading ? 'Memproses...' : 'Masuk ke Sistem' }}</span>
        </button>
      </form>

    </div>
  </div>
</template>
