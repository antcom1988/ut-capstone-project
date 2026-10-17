// @vitest-environment jsdom
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import Portal from './pegawai/PegawaiDashboardView.vue'
import Dashboard from './admin/DashboardView.vue'
import Master from './admin/MasterPegawaiView.vue'
import Berkas from './admin/BerkasAdminView.vue'
import Diklat from './pegawai/DiklatView.vue'
import ChangePassword from './ChangePasswordView.vue'
it.each([Diklat, Portal])('edits diklat retaining existing certificate and confirms deletion', async component => {
 const w = mount(component); await flushPromises(); const vm = w.vm as any
 const d = {id:5,nama_diklat:'Pelatihan',penyelenggara:'BPSDM',jumlah_jp:12,tanggal_mulai:'2026-01-01T00:00:00Z',tanggal_selesai:'2026-01-02T00:00:00Z',file_name:'certificate.pdf'}
 vm.openEditDiklat(d); await vm.submitDiklat(); expect(api.put).toHaveBeenCalledWith('/diklat/5', expect.any(FormData))
 expect((vi.mocked(api.put).mock.calls[0][1] as FormData).has('file_sertifikat')).toBe(false)
 vi.spyOn(window,'confirm').mockReturnValueOnce(false); await vm.deleteDiklat(d); expect(api.delete).not.toHaveBeenCalled()
 vi.mocked(window.confirm).mockReturnValueOnce(true); await vm.deleteDiklat(d); expect(api.delete).toHaveBeenCalledWith('/diklat/5'); w.unmount()
})
it('changes password using the backend fields, then requires reauthentication', async () => {
 const auth = useAuthStore(); auth.setAuth('test',{id:1,username:'employee',role:'PEGAWAI',must_change_password:true}); auth.isBootstrapped = true
 await router.push('/change-password'); await router.isReady()
 const w = mount(ChangePassword,{global:{plugins:[router]}}); const vm = w.vm as any
 vm.currentPassword = 'initial-pass'; vm.newPassword = 'changed-pass'; vm.confirmation = 'mismatch'; await vm.submit(); expect(api.put).not.toHaveBeenCalled()
 vm.confirmation = 'changed-pass'; await vm.submit(); expect(api.put).toHaveBeenCalledWith('/auth/password',{current_password:'initial-pass',new_password:'changed-pass'}); expect(auth.isAuthenticated).toBe(false); expect(router.currentRoute.value.path).toBe('/login'); w.unmount()
})
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

it('forces password rotation before protected navigation', async () => {
 const auth = useAuthStore(); auth.setAuth('test', {id:1,username:'employee',role:'PEGAWAI',must_change_password:true} as any); auth.isBootstrapped = true
 await router.push('/pegawai/dashboard'); expect(router.currentRoute.value.path).toBe('/change-password')
})
it('restores mandatory password state from the profile', async () => {
 const auth = useAuthStore(); auth.setAuth('test',{id:1,username:'employee',role:'PEGAWAI',nip:employee.nip})
 vi.mocked(api.get).mockResolvedValueOnce(ok({pegawai:{...employee,user_id:1},must_change_password:true}))
 await auth.bootstrap(); expect(auth.user?.must_change_password).toBe(true); expect(auth.user?.nip).toBe(employee.nip)
})
it('verifies a queue item with its exact uploaded timestamp', async () => {
 const w = mount(Berkas); await flushPromises(); const vm = w.vm as any
 vm.verifyPending({id:88,nip:employee.nip,nama_lengkap:'Other employee',nama_dokumen:'KTP',uploaded_at:'2026-01-01T00:00:00.123456Z'})
 await vm.submitVerifikasi(); expect(api.put).toHaveBeenCalledWith('/admin/berkas/verify/88',expect.objectContaining({uploaded_at:'2026-01-01T00:00:00.123456Z'})); w.unmount()
})
it('shows overdue wording for both alert audiences', async () => {
 const alert = {tipe_alert:'KGB',status:'SEGERA',sisa_hari:-20,sisa_bulan:-0.7,tanggal_jatuh_tempo:'2026-01-01'}
 vi.mocked(api.get).mockImplementation(async url => ok(url === '/pegawai/me/dashboard' ? {pegawai:employee,alerts:[alert],kelengkapan_berkas:0} : url === '/admin/dashboard' ? {alerts:[alert],berkas_lengkap_persen:0} : []))
 for (const component of [Portal,Dashboard]) { const w = mount(component); await flushPromises(); expect(w.text()).toContain('Terlambat 20 hari'); w.unmount() }
})
import api from '@/api/axios'
vi.mock('@/api/axios', () => ({ default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() } }))
const ok = (data: any) => ({ data: { success: true, data } })
const employee = { nip: '123456789012345678', nama_lengkap: 'Employee', bidang: 'Unit', golongan: 'III/a', nama_jabatan: 'Analis' }
beforeEach(() => {
 vi.clearAllMocks(); window.scrollTo = vi.fn(); localStorage.clear(); setActivePinia(createPinia())
 vi.mocked(api.get).mockImplementation(async url => ok(url === '/pegawai/me/dashboard' ? { pegawai: employee, alerts: [], kelengkapan_berkas: 25 } : url === '/admin/dashboard' ? { alerts: [], berkas_lengkap_persen: 0, threshold_months: 6 } : []))
 vi.mocked(api.post).mockResolvedValue(ok({})); vi.mocked(api.put).mockResolvedValue(ok({})); vi.mocked(api.delete).mockResolvedValue(ok({}))
})
it('refreshes completeness after a portal upload', async () => {
 const w = mount(Portal); await flushPromises(); const vm = w.vm as any
 vm.openUploadModal({ dokumen_id: 1, max_size_mb: 2 }); vm.selectedFile = new File(['pdf'], 'a.pdf', {type:'application/pdf'})
 vi.mocked(api.get).mockClear(); await vm.submitUpload(); await flushPromises()
 expect(api.get).toHaveBeenCalledWith('/pegawai/me/dashboard'); expect(w.text()).toContain('25%'); w.unmount()
})
it('keeps dashboard threshold from the backend and displays load failures', async () => {
 const w = mount(Dashboard); await flushPromises(); expect((w.vm as any).thresholdMonths).toBe(6)
 vi.mocked(api.get).mockRejectedValueOnce(new Error('offline')); await (w.vm as any).fetchDashboard(); await flushPromises()
 expect(w.find('[role="alert"]').exists()).toBe(true); w.unmount()
})
it('requires a masked unique initial password and submits canonical SMA', async () => {
 const w = mount(Master); await flushPromises(); const vm = w.vm as any; vm.openAddModal(); await flushPromises()
 vm.form.nip = employee.nip; await vm.savePegawai(); expect(api.post).not.toHaveBeenCalled()
 const password = w.find('input[autocomplete="new-password"]'); expect(password.attributes('type')).toBe('password'); expect(password.attributes('minlength')).toBe('8')
 expect(w.find('option[value="SMA"]').exists()).toBe(true); w.unmount()
})
