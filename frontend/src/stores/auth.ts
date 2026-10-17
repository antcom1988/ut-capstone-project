import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { User } from '@/types'
import api from '@/api/axios'
import { parseStoredUser } from '@/utils/validation'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('sikapj_token'))
  const user = ref<User | null>(parseStoredUser(localStorage.getItem('sikapj_user')))
  const isBootstrapped = ref(false)
  if (token.value && !user.value) localStorage.removeItem('sikapj_token')
  const isAuthenticated = computed(() => Boolean(token.value && user.value))
  const isAdmin = computed(() => user.value?.role === 'ADMIN')
  const isPegawai = computed(() => user.value?.role === 'PEGAWAI')
  function setAuth(newToken: string, newUser: User) { token.value=newToken; user.value=newUser; localStorage.setItem('sikapj_token',newToken); localStorage.setItem('sikapj_user',JSON.stringify(newUser)) }
  function logout() { token.value=null; user.value=null; localStorage.removeItem('sikapj_token'); localStorage.removeItem('sikapj_user') }
  async function bootstrap() {
    if (isBootstrapped.value || !token.value) { isBootstrapped.value=true; return }
    try {
      const res=await api.get('/auth/profile'); const data=res.data.data
      if (data?.user) setAuth(token.value, { ...data.user, nip: data.user.nip || data.pegawai?.nip || user.value?.nip, must_change_password: data.must_change_password ?? data.user.must_change_password ?? user.value?.must_change_password })
      else if (data?.pegawai) setAuth(token.value,{id:data.pegawai.user_id,username:data.pegawai.nip,role:'PEGAWAI',nip:data.pegawai.nip,must_change_password:data.must_change_password ?? user.value?.must_change_password})
    } catch { logout() } finally { isBootstrapped.value=true }
  }
  return { token,user,isAuthenticated,isAdmin,isPegawai,isBootstrapped,setAuth,logout,bootstrap }
})
