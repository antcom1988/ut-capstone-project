import type { User } from '@/types'

export const MAX_KINERJA_YEAR = new Date().getFullYear() + 1
export function validateKinerjaYear(year: number): string {
  return Number.isInteger(year) && year >= 2000 && year <= MAX_KINERJA_YEAR ? '' : `Tahun harus antara 2000 dan ${MAX_KINERJA_YEAR}.`
}

export const ALLOWED_UPLOAD_TYPES = ['application/pdf', 'image/jpeg', 'image/png']

export function validateNip(value: string): string {
  return /^\d{18}$/.test(value) ? '' : 'NIP wajib terdiri dari tepat 18 digit angka.'
}

export function validateContact(email: string, phone: string): string {
  if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return 'Format email tidak valid.'
  if (phone && !/^(?:\+62|62|0)8\d{8,12}$/.test(phone.replace(/[\s-]/g, ''))) return 'Nomor telepon Indonesia tidak valid.'
  return ''
}

export function validateDateRange(start: string, end: string): string {
  return start && end && end < start ? 'Tanggal selesai harus sama dengan atau setelah tanggal mulai.' : ''
}

export function validateFile(file: File, maxMb: number, allowedTypes = ALLOWED_UPLOAD_TYPES): string {
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const extensionMime: Record<string, string> = { pdf: 'application/pdf', jpg: 'image/jpeg', jpeg: 'image/jpeg', png: 'image/png' }
  const effectiveType = file.type || extensionMime[extension] || ''
  if (!allowedTypes.includes(effectiveType)) return 'Format file harus PDF, JPG, atau PNG.'
  if (file.size > maxMb * 1024 * 1024) return `Ukuran file melebihi batas ${maxMb} MB.`
  return ''
}

export function parseStoredUser(raw: string | null): User | null {
  if (!raw) return null
  try {
    const value = JSON.parse(raw)
    if (!value || typeof value.id !== 'number' || typeof value.username !== 'string' || !['ADMIN', 'PEGAWAI'].includes(value.role)) return null
    return value as User
  } catch { return null }
}

export function apiErrorMessage(error: any, fallback: string): string {
  if (!error?.response) return 'Tidak dapat terhubung ke server. Periksa koneksi lalu coba lagi.'
  return error.response.data?.message || fallback
}
