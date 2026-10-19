const STORAGE_KEY = 'token'

class MemoryStorage {
	private store = new Map<string, string>()
	getItem(key: string): string | null {
		return this.store.get(key) ?? null
	}
	setItem(key: string, value: string): void {
		this.store.set(key, value)
	}
	removeItem(key: string): void {
		this.store.delete(key)
	}
}

interface TokenStorage {
	getItem(key: string): string | null
	setItem(key: string, value: string): void
	removeItem(key: string): void
}

const storage: TokenStorage = ((): TokenStorage => {
	try {
		if (typeof localStorage !== 'undefined' && localStorage) {
			return localStorage as unknown as TokenStorage
		}
	} catch {
		// localStorage недоступен (например, jsdom/node окружение)
	}
	return new MemoryStorage()
})()

export function saveToken(token: string): void {
	storage.setItem(STORAGE_KEY, token)
}

export function getToken(): string | null {
	return storage.getItem(STORAGE_KEY)
}

export function clearToken(): void {
	storage.removeItem(STORAGE_KEY)
}

export function authHeaders(): { Authorization?: string } {
	const token = getToken()
	if (!token) return {}
	return { Authorization: 'Bearer ' + token }
}
