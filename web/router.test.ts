import { describe, it, expect } from 'vitest'
import { createRouter } from '../src/router'

describe('router (TST-FE-002)', () => {
	it('защищённый маршрут без токена перенаправляет на логин', () => {
		const router = createRouter({ token: null })
		expect(router.isProtected('/guests')).toBe(true)
		expect(router.redirectFor('/guests')).toBe('/login')
	})

	it('после входа защищённый маршрут доступен', () => {
		const router = createRouter({ token: 'valid-token' })
		expect(router.canAccess('/guests')).toBe(true)
	})
})
