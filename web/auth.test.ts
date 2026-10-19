import { describe, it, expect } from 'vitest'
import { saveToken, getToken, clearToken, authHeaders } from '../src/auth'

describe('auth (TST-FE-003)', () => {
	it('сохраняет токен после входа и возвращает его', () => {
		saveToken('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.token')
		expect(getToken()).toBe('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.token')
	})

	it('добавляет токен в заголовок Authorization', () => {
		saveToken('test-token-123')
		expect(authHeaders()).toEqual({ Authorization: 'Bearer test-token-123' })
	})

	it('очищает токен при выходе', () => {
		saveToken('test-token-123')
		clearToken()
		expect(getToken()).toBeNull()
	})
})
