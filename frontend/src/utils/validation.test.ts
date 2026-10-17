import { describe, expect, it } from 'vitest'
import { parseStoredUser, validateContact, validateDateRange, validateFile, validateNip } from './validation'

describe('validation helpers', () => {
  it('accepts exactly 18 numeric NIP digits', () => {
    expect(validateNip('198507152010011012')).toBe('')
    expect(validateNip('123')).toContain('18 digit')
    expect(validateNip('19850715201001101A')).toContain('18 digit')
  })
  it('validates contact email and Indonesian phone numbers', () => {
    expect(validateContact('pegawai@jemberkab.go.id', '081234567890')).toBe('')
    expect(validateContact('bukan-email', '081234567890')).toContain('email')
    expect(validateContact('pegawai@jemberkab.go.id', '123')).toContain('telepon')
  })
  it('rejects reversed date ranges', () => {
    expect(validateDateRange('2026-01-02', '2026-01-01')).toContain('setelah')
    expect(validateDateRange('2026-01-01', '2026-01-02')).toBe('')
  })
  it('checks upload MIME type and byte limit', () => {
    const valid = new File(['x'], 'scan.pdf', { type: 'application/pdf' })
    const invalid = new File(['x'], 'scan.svg', { type: 'image/svg+xml' })
    const large = new File([new Uint8Array(1025)], 'scan.pdf', { type: 'application/pdf' })
    const pdfWithoutBrowserMime = new File(['x'], 'scan.PDF', { type: '' })
    expect(validateFile(valid, 1 / 1024)).toBe('')
    expect(validateFile(pdfWithoutBrowserMime, 1, ['application/pdf'])).toBe('')
    expect(validateFile(invalid, 1)).toContain('PDF, JPG, atau PNG')
    expect(validateFile(large, 1 / 1024)).toContain('melebihi')
  })
  it('safely parses only supported persisted users', () => {
    expect(parseStoredUser('{rusak')).toBeNull()
    expect(parseStoredUser(JSON.stringify({ id: 1, username: 'admin', role: 'ROOT' }))).toBeNull()
    expect(parseStoredUser(JSON.stringify({ id: 1, username: 'admin', role: 'ADMIN' }))).toEqual({ id: 1, username: 'admin', role: 'ADMIN' })
  })
})
