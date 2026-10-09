import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import api from '@/api/axios'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('sikapj_token'))
  const user = ref<User | null>(
    localStorage.getItem('sikapj_user') ? JSON.parse(localStorage.getItem('sikapj_user')!) : null
  )

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'ADMIN')
  const isPegawai = computed(() => user.value?.role === 'PEGAWAI')

  function setAuth(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem('sikapj_token', newToken)
    localStorage.setItem('sikapj_user', JSON.stringify(newUser))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('sikapj_token')
    localStorage.removeItem('sikapj_user')
  }

  async function fetchProfile() {
    try {
      const res = await api.get('/auth/profile')
      if (res.data.success) {
        // update user
        if (res.data.data.pegawai) {
          user.value = {
            id: res.data.data.pegawai.user_id,
            username: res.data.data.pegawai.nip,
            role: 'PEGAWAI',
            nip: res.data.data.pegawai.nip,
          }
        }
      }
    } catch (e) {
      console.error(e)
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    isPegawai,
    setAuth,
    logout,
    fetchProfile,
  }
})
