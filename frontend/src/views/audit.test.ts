// @vitest-environment jsdom
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import Berkas from './admin/BerkasAdminView.vue'
import Diklat from './pegawai/DiklatView.vue'
import Kinerja from './pegawai/KinerjaView.vue'
import Portal from './pegawai/PegawaiDashboardView.vue'
import api from '@/api/axios'
vi.mock('@/api/axios', () => ({ default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() } }))
it.each([Kinerja, Portal])('accepts image kinerja and rejects oversized files and out-of-range years in both forms', async (component) => {
 const w = mount(component); await flushPromises(); const vm = w.vm as any
 const form = vm.form || vm.kinerjaForm
 form.tahun_periode = new Date().getFullYear() + 1
 vm.kinerjaFile = new File(['image'], 'scan.jpg', { type: 'image/jpeg' })
 await vm.submitKinerja()
 expect(api.post).toHaveBeenCalledWith('/kinerja/upload', expect.any(FormData))
 vi.mocked(api.post).mockClear()
 vm.kinerjaFile = new File([new Uint8Array(1048577)], 'scan.pdf', { type: 'application/pdf' })
 await vm.submitKinerja(); expect(api.post).not.toHaveBeenCalled()
 vm.kinerjaFile = new File(['pdf'], 'scan.pdf', { type: 'application/pdf' })
 form.tahun_periode = new Date().getFullYear() + 2
 await vm.submitKinerja(); expect(api.post).not.toHaveBeenCalled()
 w.unmount()
})
it.each([Diklat, Portal])('requires certificate evidence before creating diklat in both input paths', async (component) => {
 const w = mount(component); await flushPromises(); const vm = w.vm as any
 if (vm.openModal) vm.openModal(); else vm.openAddDiklatModal()
 await vm.submitDiklat(); await flushPromises()
 expect(api.post).not.toHaveBeenCalled()
 expect(w.text()).toContain('Sertifikat wajib diunggah')
 w.unmount()
})
const ok = (data: any) => ({ data: { success: true, data } })
const employee = { nip: '123456789012345678', nama_lengkap: 'Employee', bidang: 'Unit', golongan: 'III/a', nama_jabatan: 'Analis' }
beforeEach(() => { vi.clearAllMocks(); localStorage.clear(); setActivePinia(createPinia()); vi.mocked(api.get).mockImplementation(async (url) => ok(url === '/admin/pegawai' ? [employee] : url === '/pegawai/me/dashboard' ? { pegawai: employee, alerts: [], kelengkapan_berkas: 0 } : [])); vi.mocked(api.put).mockResolvedValue(ok({})); vi.mocked(api.post).mockResolvedValue(ok({})) })
it('verifies the fetched upload version and displays stale conflicts without dismissing the decision', async () => {
 const w = mount(Berkas); await flushPromises()
 const vm = w.vm as any
 vm.openVerifyModal({ berkas_id: 7, nama_dokumen: 'KTP', uploaded_at: '2026-01-01T00:00:00Z' })
 vi.mocked(api.put).mockRejectedValueOnce({ response: { status: 409, data: { message: 'Berkas berubah. Muat ulang.' } } })
 await vm.submitVerifikasi(); await flushPromises()
 expect(api.put).toHaveBeenCalledWith('/admin/berkas/verify/7', expect.objectContaining({ uploaded_at: '2026-01-01T00:00:00Z' }))
 expect(w.text()).toContain('Berkas berubah. Muat ulang.')
 expect(w.text()).toContain('Simpan Keputusan')
 w.unmount()
})
