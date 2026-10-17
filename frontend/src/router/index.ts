import { createRouter,createWebHistory,type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
const routes:RouteRecordRaw[]=[
 {path:'/change-password',name:'change-password',component:()=>import('@/views/ChangePasswordView.vue'),meta:{requiresAuth:true}},
 {path:'/login',name:'login',component:()=>import('@/views/LoginView.vue'),meta:{guestOnly:true}},
 {path:'/',component:()=>import('@/layouts/AppLayout.vue'),meta:{requiresAuth:true},children:[
  {path:'',redirect:()=>useAuthStore().isAdmin?'/admin/dashboard':'/pegawai/dashboard'},
  {path:'admin/dashboard',name:'admin-dashboard',component:()=>import('@/views/admin/DashboardView.vue'),meta:{role:'ADMIN'}},
  {path:'admin/berkas',name:'admin-berkas',component:()=>import('@/views/admin/BerkasAdminView.vue'),meta:{role:'ADMIN'}},
  {path:'admin/pegawai',name:'admin-pegawai',component:()=>import('@/views/admin/MasterPegawaiView.vue'),meta:{role:'ADMIN'}},
  {path:'admin/laporan',name:'admin-laporan',component:()=>import('@/views/admin/ReportsView.vue'),meta:{role:'ADMIN'}},
  {path:'admin/audit-log',name:'admin-audit',component:()=>import('@/views/admin/AuditLogView.vue'),meta:{role:'ADMIN'}},
  {path:'pegawai/dashboard',name:'pegawai-dashboard',component:()=>import('@/views/pegawai/PegawaiDashboardView.vue'),meta:{role:'PEGAWAI'}},
  {path:'pegawai/diklat',name:'pegawai-diklat',component:()=>import('@/views/pegawai/DiklatView.vue'),meta:{role:'PEGAWAI'}},
  {path:'pegawai/kinerja',name:'pegawai-kinerja',component:()=>import('@/views/pegawai/KinerjaView.vue'),meta:{role:'PEGAWAI'}}]},
 {path:'/:pathMatch(.*)*',name:'not-found',component:()=>import('@/views/NotFoundView.vue')}
]
const router=createRouter({history:createWebHistory(),routes,scrollBehavior:()=>({top:0})})
router.beforeEach(async to=>{const auth=useAuthStore();await auth.bootstrap();if(to.meta.requiresAuth&&!auth.isAuthenticated)return{name:'login',query:{redirect:to.fullPath}};if(auth.isAuthenticated&&auth.user?.must_change_password&&to.name!=='change-password')return{name:'change-password'};if(to.meta.guestOnly&&auth.isAuthenticated)return{name:auth.isAdmin?'admin-dashboard':'pegawai-dashboard'};if(to.meta.role&&auth.user?.role!==to.meta.role)return{name:auth.isAdmin?'admin-dashboard':'pegawai-dashboard'};})
export default router
