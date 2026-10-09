import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/LoginView.vue'
import AppLayout from '@/layouts/AppLayout.vue'
import DashboardView from '@/views/admin/DashboardView.vue'
import BerkasAdminView from '@/views/admin/BerkasAdminView.vue'
import MasterPegawaiView from '@/views/admin/MasterPegawaiView.vue'
import PegawaiDashboardView from '@/views/pegawai/PegawaiDashboardView.vue'
import DiklatView from '@/views/pegawai/DiklatView.vue'
import KinerjaView from '@/views/pegawai/KinerjaView.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { guestOnly: true },
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: () => {
          const auth = useAuthStore()
          return auth.isAdmin ? '/admin/dashboard' : '/pegawai/dashboard'
        },
      },
      // Admin Routes
      {
        path: 'admin/dashboard',
        name: 'admin-dashboard',
        component: DashboardView,
        meta: { role: 'ADMIN' },
      },
      {
        path: 'admin/berkas',
        name: 'admin-berkas',
        component: BerkasAdminView,
        meta: { role: 'ADMIN' },
      },
      {
        path: 'admin/pegawai',
        name: 'admin-pegawai',
        component: MasterPegawaiView,
        meta: { role: 'ADMIN' },
      },
      // Pegawai Routes
      {
        path: 'pegawai/dashboard',
        name: 'pegawai-dashboard',
        component: PegawaiDashboardView,
        meta: { role: 'PEGAWAI' },
      },
      {
        path: 'pegawai/diklat',
        name: 'pegawai-diklat',
        component: DiklatView,
        meta: { role: 'PEGAWAI' },
      },
      {
        path: 'pegawai/kinerja',
        name: 'pegawai-kinerja',
        component: KinerjaView,
        meta: { role: 'PEGAWAI' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return next({ name: 'login' })
  }

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    if (authStore.isAdmin) return next({ name: 'admin-dashboard' })
    return next({ name: 'pegawai-dashboard' })
  }

  if (to.meta.role && authStore.user?.role !== to.meta.role) {
    if (authStore.isAdmin) return next({ name: 'admin-dashboard' })
    return next({ name: 'pegawai-dashboard' })
  }

  next()
})

export default router
