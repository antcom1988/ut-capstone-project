<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/axios'
import { apiErrorMessage } from '@/utils/validation'
const auth = useAuthStore(), router = useRouter()
const currentPassword = ref(''), newPassword = ref(''), confirmation = ref(''), error = ref(''), saving = ref(false)
async function submit() {
 error.value = ''
 if (!currentPassword.value || newPassword.value.length < 10) { error.value = 'Isi kata sandi saat ini dan kata sandi baru minimal 10 karakter.'; return }
 if (newPassword.value !== confirmation.value) { error.value = 'Konfirmasi kata sandi tidak cocok.'; return }
 if (currentPassword.value === newPassword.value) { error.value = 'Kata sandi baru harus berbeda.'; return }
 saving.value = true
 try {
  await api.put('/auth/password', {current_password: currentPassword.value, new_password: newPassword.value})
  auth.logout()
  sessionStorage.setItem('sikapj_notice', 'Kata sandi berhasil diubah. Silakan masuk kembali.')
  await router.replace('/login')
 } catch (e) { error.value = apiErrorMessage(e, 'Gagal mengganti kata sandi.') }
 finally { saving.value = false }
}
function logout() { auth.logout(); router.replace('/login') }
</script>
<template>
 <main class="min-h-screen bg-slate-100 flex items-center justify-center p-6">
  <form @submit.prevent="submit" class="bg-white rounded-xl shadow p-6 space-y-4 w-full max-w-md">
   <h1 class="text-xl font-bold">Ganti Kata Sandi</h1>
   <p v-if="auth.user?.must_change_password" class="text-sm text-amber-800">Anda wajib mengganti kata sandi awal sebelum mengakses sistem.</p>
   <p v-if="error" role="alert" class="text-rose-700">{{ error }}</p>
   <label class="block">Kata sandi saat ini<input v-model="currentPassword" required type="password" autocomplete="current-password" class="block border rounded p-2 w-full" /></label>
   <label class="block">Kata sandi baru (minimal 10 karakter)<input v-model="newPassword" required minlength="10" type="password" autocomplete="new-password" class="block border rounded p-2 w-full" /></label>
   <label class="block">Konfirmasi kata sandi baru<input v-model="confirmation" required minlength="10" type="password" autocomplete="new-password" class="block border rounded p-2 w-full" /></label>
   <button :disabled="saving" class="bg-primary-600 text-white rounded p-2">{{ saving ? 'Menyimpan...' : 'Simpan Kata Sandi' }}</button>
   <button type="button" @click="logout" class="ml-3">Keluar</button>
  </form>
 </main>
</template>
