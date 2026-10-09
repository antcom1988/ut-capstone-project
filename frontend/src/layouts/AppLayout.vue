<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen flex bg-slate-100 text-slate-800">
    <!-- Sidebar -->
    <aside class="w-64 bg-primary-700 text-white flex flex-col justify-between hidden md:flex shadow-xl">
      <div>
        <!-- Logo & Branding -->
        <div class="p-6 border-b border-primary-600/50 flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-white/10 flex items-center justify-center font-bold text-lg text-white">
            SJ
          </div>
          <div>
            <div class="font-bold text-base tracking-tight leading-tight">SIKAP - J</div>
            <div class="text-[10px] text-primary-200 tracking-wider uppercase font-semibold">BPKAD Jember</div>
          </div>
        </div>

        <!-- Nav Items -->
        <nav class="p-4 space-y-1.5 text-sm">
          <!-- Menu Admin -->
          <template v-if="authStore.isAdmin">
            <div class="text-[11px] font-bold text-primary-300 uppercase tracking-wider px-3 py-2">Menu Utama</div>
            <router-link
              to="/admin/dashboard"
              active-class="bg-white/15 text-white font-semibold"
              class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-primary-100 hover:bg-white/10 transition-colors"
            >
              <span>📊</span>
              <span>Dashboard Monitoring</span>
            </router-link>
            <router-link
              to="/admin/berkas"
              active-class="bg-white/15 text-white font-semibold"
              class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-primary-100 hover:bg-white/10 transition-colors"
            >
              <span>📁</span>
              <span>Verifikasi E-Berkas</span>
            </router-link>
            <router-link
              to="/admin/pegawai"
              active-class="bg-white/15 text-white font-semibold"
              class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-primary-100 hover:bg-white/10 transition-colors"
            >
              <span>👥</span>
              <span>Master Data Pegawai</span>
            </router-link>
          </template>

          <!-- Menu Pegawai -->
          <template v-if="authStore.isPegawai">
            <div class="text-[11px] font-bold text-primary-300 uppercase tracking-wider px-3 py-2">Portal Mandiri</div>
            <router-link
              to="/pegawai/dashboard"
              active-class="bg-white/15 text-white font-semibold"
              class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-primary-100 hover:bg-white/10 transition-colors"
            >
              <span>👤</span>
              <span>Profil & E-Berkas</span>
            </router-link>
            <router-link
              to="/pegawai/diklat"
              active-class="bg-white/15 text-white font-semibold"
              class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-primary-100 hover:bg-white/10 transition-colors"
            >
              <span>🎓</span>
              <span>Rekam Jejak Diklat</span>
            </router-link>
            <router-link
              to="/pegawai/kinerja"
              active-class="bg-white/15 text-white font-semibold"
              class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-primary-100 hover:bg-white/10 transition-colors"
            >
              <span>📊</span>
              <span>Arsip Kinerja SKP/PK</span>
            </router-link>
          </template>
        </nav>
      </div>

      <!-- User Info & Logout -->
      <div class="p-4 border-t border-primary-600/50 bg-primary-800/40">
        <div class="flex items-center justify-between">
          <div class="truncate pr-2">
            <div class="text-xs font-bold text-white truncate">{{ authStore.user?.username }}</div>
            <div class="text-[10px] text-primary-200 uppercase font-semibold">{{ authStore.user?.role }}</div>
          </div>
          <button
            @click="handleLogout"
            class="p-2 text-primary-200 hover:text-white hover:bg-white/10 rounded-lg transition-colors text-xs font-medium"
            title="Keluar"
          >
            Keluar 🚪
          </button>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Top Mobile Bar -->
      <header class="md:hidden bg-primary-700 text-white p-4 flex items-center justify-between">
        <div class="font-bold text-base">SIKAP - J</div>
        <button @click="handleLogout" class="text-xs bg-white/20 px-2.5 py-1 rounded">Keluar</button>
      </header>

      <main class="flex-1 p-6 md:p-8 max-w-7xl mx-auto w-full">
        <router-view />
      </main>
    </div>
  </div>
</template>
